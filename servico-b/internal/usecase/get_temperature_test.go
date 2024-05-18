package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/regismartiny/desafio-cloudrun/internal/viacep"
	"github.com/regismartiny/desafio-cloudrun/internal/weatherapi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mocks
type CepValidatorMock struct {
	mock.Mock
}

func (m *CepValidatorMock) IsCEP(cep string) bool {
	args := m.Called(cep)
	return args.Bool(0)
}

type ViaCepClientMock struct {
	mock.Mock
}

func (m *ViaCepClientMock) GetAddressInfo(ctx *context.Context, cep string) (viacep.Address, error) {
	args := m.Called(ctx, cep)
	return args.Get(0).(viacep.Address), args.Error(1)
}

type WeatherApiClientMock struct {
	mock.Mock
}

func (m *WeatherApiClientMock) GetWeatherInfo(ctx *context.Context, cidade string) (weatherapi.Weather, error) {
	args := m.Called(ctx, cidade)
	return args.Get(0).(weatherapi.Weather), args.Error(1)
}

// Tests
type GetTemperatureUseCaseTestSuite struct {
	suite.Suite
	CepValidator     CepValidatorMock
	ViaCepClient     ViaCepClientMock
	WeatherApiClient WeatherApiClientMock
}

func (suite *GetTemperatureUseCaseTestSuite) SetupTest() {
	suite.CepValidator = CepValidatorMock{}
	suite.ViaCepClient = ViaCepClientMock{}
	suite.WeatherApiClient = WeatherApiClientMock{}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(GetTemperatureUseCaseTestSuite))
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenValidCepProvided_ThenShouldReturnTemperatureInfo() {
	// arrange
	addressMock := viacep.Address{Localidade: "São Paulo"}
	weatherMock := weatherapi.Weather{Current: weatherapi.Current{TempC: 20, TempF: 68}}

	suite.CepValidator.On("IsCEP", "95770000").Return(true)
	suite.ViaCepClient.On("GetAddressInfo", mock.Anything, "95770000").Return(addressMock, nil)
	suite.WeatherApiClient.On("GetWeatherInfo", mock.Anything, "São Paulo").Return(weatherMock, nil)

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.ViaCepClient, &suite.WeatherApiClient)
	output, err := useCase.Execute("95770000")

	// assert
	suite.NotNil(output)
	suite.Nil(err)
	suite.Equal(20.0, output.TempC)
	suite.Equal(68.0, output.TempF)
	suite.Equal(293.0, output.TempK)
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenInvalidCepProvided_ThenShouldReturnErrorInvalidZipcode() {
	// arrange
	suite.CepValidator.On("IsCEP", "1234").Return(false)

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.ViaCepClient, &suite.WeatherApiClient)
	output, err := useCase.Execute("1234")

	// assert
	suite.NotNil(output)
	suite.NotNil(err)
	suite.Equal(GetTemperatureOutputDTO{}, output)
	suite.Equal("invalid zipcode", err.Error())
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenValidCepProvidedButNotFoundInViacep_ThenShouldReturnErrorCannotFindZipcode() {
	// arrange
	addressMock := viacep.Address{}

	suite.CepValidator.On("IsCEP", "99999999").Return(true)
	suite.ViaCepClient.On("GetAddressInfo", mock.Anything, "99999999").Return(addressMock, errors.New("can not find zipcode"))

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.ViaCepClient, &suite.WeatherApiClient)
	output, err := useCase.Execute("99999999")

	// assert
	suite.NotNil(output)
	suite.NotNil(err)
	suite.Equal(GetTemperatureOutputDTO{}, output)
	suite.Equal("can not find zipcode", err.Error())
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenValidCepProvidedButNotFoundInWeatherAPI_ThenShouldReturnErrorCannotFindZipcode() {
	// arrange
	addressMock := viacep.Address{Localidade: "São Paulo"}
	weatherMock := weatherapi.Weather{}

	suite.CepValidator.On("IsCEP", "99999999").Return(true)
	suite.ViaCepClient.On("GetAddressInfo", mock.Anything, "99999999").Return(addressMock, nil)
	suite.WeatherApiClient.On("GetWeatherInfo", mock.Anything, "São Paulo").Return(weatherMock, errors.New("can not find zipcode"))

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.ViaCepClient, &suite.WeatherApiClient)
	output, err := useCase.Execute("99999999")

	// assert
	suite.NotNil(output)
	suite.NotNil(err)
	suite.Equal(GetTemperatureOutputDTO{}, output)
	suite.Equal("can not find zipcode", err.Error())
}
