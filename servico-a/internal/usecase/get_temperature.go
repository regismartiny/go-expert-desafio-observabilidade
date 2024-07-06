package usecase

import (
	"context"
	"errors"
	"log"

	temperature "servico-a/internal/temperature"
)

type GetTemperatureOutputDTO struct {
	City  string  `json:"city"`   // Nome da Cidade
	TempC float64 `json:"temp_C"` // Celsius
	TempF float64 `json:"temp_F"` // Fahrenheit
	TempK float64 `json:"temp_K"` // Kelvin
}

type CepValidator interface {
	IsCEP(input string) bool
}

type TemperatureClient interface {
	GetTemperatureInfo(ctx *context.Context, cep string) (temperature.TemperatureInfo, error)
}

type GetTemperatureUseCase struct {
	CepValidator      CepValidator
	TemperatureClient TemperatureClient
}

func NewGetTemperatureUseCase(
	CepValidator CepValidator,
	TemperatureClient TemperatureClient,
) *GetTemperatureUseCase {
	return &GetTemperatureUseCase{
		CepValidator:      CepValidator,
		TemperatureClient: TemperatureClient,
	}
}

func (c *GetTemperatureUseCase) Execute(context *context.Context, input string) (GetTemperatureOutputDTO, error) {

	log.Println("Validating CEP", input)

	if !c.CepValidator.IsCEP(input) {
		return GetTemperatureOutputDTO{}, errors.New("invalid zipcode")
	}

	temperatureInfo, err := getTemperatureInfo(context, c.TemperatureClient, input)

	if err != nil {
		return GetTemperatureOutputDTO{}, errors.New("can not find zipcode")
	}

	return GetTemperatureOutputDTO{
		City:  temperatureInfo.City,
		TempC: temperatureInfo.TempC,
		TempF: temperatureInfo.TempF,
		TempK: convertCelsiusToKelvin(temperatureInfo.TempC),
	}, nil

}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func getTemperatureInfo(ctx *context.Context, client TemperatureClient, cep string) (temperature.TemperatureInfo, error) {
	log.Println("Calling Temperature API for CEP:", cep)

	temperatureInfo, err := client.GetTemperatureInfo(ctx, cep)
	if err != nil {
		log.Println(err)
		return temperature.TemperatureInfo{}, err
	}

	log.Println("Temperature info:", temperatureInfo)

	return temperatureInfo, nil
}
