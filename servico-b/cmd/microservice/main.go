package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/regismartiny/desafio-cloudrun/configs"
	"github.com/regismartiny/desafio-cloudrun/internal/usecase"
	"github.com/regismartiny/desafio-cloudrun/internal/validators"
	"github.com/regismartiny/desafio-cloudrun/internal/viacep"
	"github.com/regismartiny/desafio-cloudrun/internal/weatherapi"
)

type HandlerData struct {
	GetTemperatureUseCase *usecase.GetTemperatureUseCase
}

func main() {

	config, _ := configs.LoadConfig(".")

	cepValidator := validators.NewCepValidator()
	viaCepClient := getViaCepClient(config.ViaCepAPIBaseURL, config.ViaCepAPIToken)
	weatherApiClient := getWeatherClient(config.WeatherAPIBaseURL, config.WeatherAPIToken)

	h := &HandlerData{
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(cepValidator, viaCepClient, weatherApiClient),
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
