package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"servico-a/configs"
	temperature "servico-a/internal/temperature"
	"servico-a/internal/usecase"
	"servico-a/internal/validators"
)

type HandlerData struct {
	GetTemperatureUseCase *usecase.GetTemperatureUseCase
}

type RequestBody struct {
	Cep string `json:"cep"`
}

func main() {

	config, _ := configs.LoadConfig(".")

	cepValidator := validators.NewCepValidator()
	temperatureClient := getTemperatureClient(config.TemperatureAPIBaseURL, config.TemperatureAPIToken)

	h := &HandlerData{
		GetTemperatureUseCase: usecase.NewGetTemperatureUseCase(cepValidator, temperatureClient),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", http.HandlerFunc(h.handlePost))

	server := &http.Server{
		Addr:    ":8070",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func getTemperatureClient(baseURLStr string, apiToken string) *temperature.Client {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}
	return temperature.NewClient(baseURL, apiToken)
}

func (h *HandlerData) handlePost(w http.ResponseWriter, r *http.Request) {

	log.Println("Post request received")

	var requestBody RequestBody

	if err := decodeRequestBody(r, &requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	output, err := h.GetTemperatureUseCase.Execute(requestBody.Cep)
	if err != nil {
		statusCode := getStatusCode(err.Error())

		w.WriteHeader(statusCode)
		w.Write([]byte(err.Error()))
		return
	}

	bytes, err := json.Marshal(output)
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
