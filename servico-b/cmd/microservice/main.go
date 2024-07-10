package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"servico-b/configs"

	"servico-b/internal/telemetry"
	"servico-b/internal/usecase"
	"servico-b/internal/viacep"
	"servico-b/internal/weatherapi"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type HandlerData struct {
	tracer                *trace.Tracer
	GetTemperatureUseCase *usecase.GetTemperatureUseCase
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := telemetry.SetupOTelSDK(ctx, "servico-b")
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	tracer := otel.Tracer("servico-b")

	mux := http.NewServeMux()

	config, _ := configs.LoadConfig(".")
	viaCepClient := getViaCepClient(config.ViaCepAPIBaseURL, config.ViaCepAPIToken)
	weatherApiClient := getWeatherClient(config.WeatherAPIBaseURL, config.WeatherAPIToken)

	h := &HandlerData{
		tracer:                &tracer,
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(&tracer, viaCepClient, weatherApiClient),
	}

	// Register handlers.
	mux.Handle("GET /temperatura/{cep}", http.HandlerFunc(h.handleGet))

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func getViaCepClient(baseURLStr string, apiToken string) *viacep.Client {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}
	return viacep.NewClient(baseURL, apiToken)
}

func getWeatherClient(baseURLStr string, apiToken string) *weatherapi.Client {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}
	return weatherapi.NewClient(baseURL, apiToken)
}

func (h *HandlerData) handleGet(w http.ResponseWriter, r *http.Request) {

	log.Println("Request received")

	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := *h.tracer
	ctx, span := tracer.Start(ctx, "get-temperature")
	defer span.End()

	cep := r.PathValue("cep")

	output, err := h.GetTemperatureUseCase.Execute(&ctx, cep)
	if err != nil {
		statusCode := getStatusCode(err.Error())

		w.WriteHeader(statusCode)
		w.Write([]byte(err.Error()))
		return
	}

	log.Println(output)

	bytes, err := json.Marshal(output)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte(bytes))
}

func getStatusCode(errorMsg string) int {
	switch errorMsg {
	case "invalid zipcode":
		return http.StatusUnprocessableEntity
	case "can not find zipcode":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
