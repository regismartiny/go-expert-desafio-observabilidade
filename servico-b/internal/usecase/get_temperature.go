package usecase

import (
	"context"
	"errors"
	"log"

	"servico-b/internal/viacep"
	"servico-b/internal/weatherapi"

	"go.opentelemetry.io/otel/trace"
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
	tracer           *trace.Tracer
	ViaCepClient     ViaCepClient
	WeatherApiClient WeatherApiClient
}

func NewGetTemperatureUseCase(
	tracer *trace.Tracer,
	ViaCepClient ViaCepClient,
	WeatherApiClient WeatherApiClient,
) *GetTemperatureUseCase {
	return &GetTemperatureUseCase{
		tracer:           tracer,
		ViaCepClient:     ViaCepClient,
		WeatherApiClient: WeatherApiClient,
	}
}

func (c *GetTemperatureUseCase) Execute(context *context.Context, input string) (GetTemperatureOutputDTO, error) {

	tracer := *c.tracer
	ctx, spanCep := tracer.Start(*context, "get-cep-info")

	addressInfo, err := getViaCepAddressInfo(&ctx, c.ViaCepClient, input)

	if err != nil {
		return GetTemperatureOutputDTO{}, errors.New("can not find zipcode")
	}

	spanCep.End()

	city := addressInfo.Localidade
	log.Println("City: " + city)

	ctx, spanWeather := tracer.Start(*context, "get-weather-info")

	weatherInfo, err := getWeatherApiInfo(&ctx, c.WeatherApiClient, city)

	if err != nil {
		return GetTemperatureOutputDTO{}, errors.New("can not find zipcode")
	}

	spanWeather.End()

	return GetTemperatureOutputDTO{
		City:  city,
		TempC: weatherInfo.Current.TempC,
		TempF: weatherInfo.Current.TempF,
		TempK: convertCelsiusToKelvin(weatherInfo.Current.TempC),
	}, nil

}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func getViaCepAddressInfo(ctx *context.Context, client ViaCepClient, cep string) (viacep.Address, error) {
	log.Println("Calling ViaCEP API for CEP:", cep)

	adrressInfo, err := client.GetAddressInfo(ctx, cep)
	if err != nil {
		log.Println(err)
		return viacep.Address{}, err
	}

	log.Println("Address info: ", adrressInfo)

	return adrressInfo, nil
}

func getWeatherApiInfo(ctx *context.Context, client WeatherApiClient, city string) (weatherapi.Weather, error) {
	log.Println("Calling Weather API for city:", city)

	weatherInfo, err := client.GetWeatherInfo(ctx, city)
	if err != nil {
		log.Println(err)
		return weatherapi.Weather{}, err
	}

	log.Println("Weather info: ", weatherInfo)

	return weatherInfo, nil
}
