package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/regismartiny/desafio-cloudrun/internal/viacep"
	"github.com/regismartiny/desafio-cloudrun/internal/weatherapi"
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

type ViaCepClient interface {
	GetAddressInfo(ctx *context.Context, cep string) (viacep.Address, error)
}

type WeatherApiClient interface {
	GetWeatherInfo(ctx *context.Context, cidade string) (weatherapi.Weather, error)
}

type GetTemperatureUseCase struct {
	ViaCepClient     ViaCepClient
	WeatherApiClient WeatherApiClient
}

func NewGetTemperatureUseCase(
	ViaCepClient ViaCepClient,
	WeatherApiClient WeatherApiClient,
) *GetTemperatureUseCase {
	return &GetTemperatureUseCase{
		ViaCepClient:     ViaCepClient,
		WeatherApiClient: WeatherApiClient,
	}
}

func (c *GetTemperatureUseCase) Execute(input string) (GetTemperatureOutputDTO, error) {
	context := context.Background()

	addressInfo, err := getViaCepAddressInfo(&context, c.ViaCepClient, input)

	cidade := addressInfo.Localidade
	log.Println("Cidade: " + cidade)

	if err != nil {
		return GetTemperatureOutputDTO{}, errors.New("can not find zipcode")
	}

	weatherInfo, err := getWeatherApiInfo(&context, c.WeatherApiClient, cidade)

	if err != nil {
		return GetTemperatureOutputDTO{}, errors.New("can not find zipcode")
	}

	return GetTemperatureOutputDTO{
		City:  cidade,
		TempC: weatherInfo.Current.TempC,
		TempF: weatherInfo.Current.TempF,
		TempK: convertCelsiusToKelvin(weatherInfo.Current.TempC),
	}, nil

}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func getViaCepAddressInfo(ctx *context.Context, client ViaCepClient, cep string) (viacep.Address, error) {
	log.Println("Searching for CEP info on ViaCEP API")

	adrressInfo, err := client.GetAddressInfo(ctx, cep)
	if err != nil {
		log.Println(err)
		return viacep.Address{}, err
	}

	log.Println("Address info: ", adrressInfo)

	return adrressInfo, nil
}

func getWeatherApiInfo(ctx *context.Context, client WeatherApiClient, cidade string) (weatherapi.Weather, error) {
	log.Println("Searching for weather info on Weather API")

	weatherInfo, err := client.GetWeatherInfo(ctx, cidade)
	if err != nil {
		log.Println(err)
		return weatherapi.Weather{}, err
	}

	log.Println("Weather info: ", weatherInfo)

	return weatherInfo, nil
}
