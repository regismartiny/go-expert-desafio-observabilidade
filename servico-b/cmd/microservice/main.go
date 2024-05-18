package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"servico-b/configs"

	"servico-b/internal/usecase"
	"servico-b/internal/viacep"
	"servico-b/internal/weatherapi"

	otel "servico-b/internal/otel"
)

type HandlerData struct {
	GetTemperatureUseCase *usecase.GetTemperatureUseCase
}

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config, _ := configs.LoadConfig(".")

	setupOpenTelemetry(ctx)

	viaCepClient := getViaCepClient(config.ViaCepAPIBaseURL, config.ViaCepAPIToken)
	weatherApiClient := getWeatherClient(config.WeatherAPIBaseURL, config.WeatherAPIToken)

	h := &HandlerData{
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(viaCepClient, weatherApiClient),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /temperatura/{cep}", http.HandlerFunc(h.handleGet))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func setupOpenTelemetry(ctx context.Context) {
	// Set up OpenTelemetry.
	otelShutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
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
