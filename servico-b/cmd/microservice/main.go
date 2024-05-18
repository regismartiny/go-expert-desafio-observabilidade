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

	"servico-b/internal/usecase"
	"servico-b/internal/viacep"
	"servico-b/internal/weatherapi"

	otel "servico-b/internal/otel"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HandlerData struct {
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
	otelShutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
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

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	config, _ := configs.LoadConfig(".")
	viaCepClient := getViaCepClient(config.ViaCepAPIBaseURL, config.ViaCepAPIToken)
	weatherApiClient := getWeatherClient(config.WeatherAPIBaseURL, config.WeatherAPIToken)

	h := &HandlerData{
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(viaCepClient, weatherApiClient),
	}

	// Register handlers.
	handleFunc("GET /temperatura/{cep}", http.HandlerFunc(h.handleGet))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
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

	cep := r.PathValue("cep")

	output, err := h.GetTemperatureUseCase.Execute(cep)
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
