package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"servico-a/configs"
	"servico-a/internal/telemetry"
	temperature "servico-a/internal/temperature"
	"servico-a/internal/usecase"
	"servico-a/internal/validators"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type HandlerData struct {
	tracer                *trace.Tracer
	GetTemperatureUseCase *usecase.GetTemperatureUseCase
}

type RequestBody struct {
	Cep string `json:"cep"`
}

func main() {

	config, _ := configs.LoadConfig(".")

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := telemetry.SetupOTelSDK(ctx, "servico-a")
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	tracer := otel.Tracer("servico-a")

	cepValidator := validators.NewCepValidator()
	temperatureClient := getTemperatureClient(&tracer, config.TemperatureAPIBaseURL, config.TemperatureAPIToken)

	h := &HandlerData{
		tracer:                &tracer,
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(cepValidator, temperatureClient),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", http.HandlerFunc(h.handlePost))

	server := &http.Server{
		Addr:        ":8070",
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- server.ListenAndServe()
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
	err = server.Shutdown(context.Background())
}

func getTemperatureClient(tracer *trace.Tracer, baseURLStr string, apiToken string) *temperature.Client {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}
	return temperature.NewClient(tracer, baseURL, apiToken)
}

func (h *HandlerData) handlePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Post request received")

	ctx := r.Context()
	tracer := *h.tracer
	ctx, span := tracer.Start(ctx, "get-temperature")
	defer span.End()

	var requestBody RequestBody

	if err := decodeRequestBody(r, &requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	output, err := h.GetTemperatureUseCase.Execute(&ctx, requestBody.Cep)
	if err != nil {
		statusCode := getStatusCode(err.Error())

		w.WriteHeader(statusCode)
		w.Write([]byte(err.Error()))
		return
	}

	writeJsonOutput(output, w)
}
func writeJsonOutput(object interface{}, w http.ResponseWriter) {
	bytes, err := json.Marshal(object)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte(bytes))
}

func decodeRequestBody(r *http.Request, requestBody *RequestBody) error {
	dec := json.NewDecoder(r.Body)
	for {
		if err := dec.Decode(requestBody); err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error decoding request body: ", err)
			return err
		}
	}
	return nil
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
