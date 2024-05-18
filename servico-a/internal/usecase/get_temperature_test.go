package usecase

import (
	"context"
	"errors"
	"testing"

	temperature "servico-a/internal/temperature"

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

type TemperatureClientMock struct {
	mock.Mock
}

func (m *TemperatureClientMock) GetTemperatureInfo(ctx *context.Context, cep string) (temperature.TemperatureInfo, error) {
	args := m.Called(ctx, cep)
	return args.Get(0).(temperature.TemperatureInfo), args.Error(1)
}

// Tests
type GetTemperatureUseCaseTestSuite struct {
	suite.Suite
	CepValidator      CepValidatorMock
	TemperatureClient TemperatureClientMock
}

func (suite *GetTemperatureUseCaseTestSuite) SetupTest() {
	suite.CepValidator = CepValidatorMock{}
	suite.TemperatureClient = TemperatureClientMock{}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(GetTemperatureUseCaseTestSuite))
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenValidCepProvided_ThenShouldReturnTemperatureInfo() {
	// arrange
	temperatureMock := temperature.TemperatureInfo{City: "Feliz", TempC: 20, TempF: 68}

	suite.CepValidator.On("IsCEP", "95770000").Return(true)
	suite.TemperatureClient.On("GetTemperatureInfo", mock.Anything, "95770000").Return(temperatureMock, nil)

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.TemperatureClient)
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
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.TemperatureClient)
	output, err := useCase.Execute("1234")

	// assert
	suite.NotNil(output)
	suite.NotNil(err)
	suite.Equal(GetTemperatureOutputDTO{}, output)
	suite.Equal("invalid zipcode", err.Error())
}

func (suite *GetTemperatureUseCaseTestSuite) TestGetTemperature_WhenValidCepProvidedButNotFoundInTemperatureAPI_ThenShouldReturnErrorCannotFindZipcode() {
	// arrange
	temperatureMock := temperature.TemperatureInfo{}

	suite.CepValidator.On("IsCEP", "99999999").Return(true)
	suite.TemperatureClient.On("GetTemperatureInfo", mock.Anything, "99999999").Return(temperatureMock, errors.New("can not find zipcode"))

	// act
	useCase := NewGetTemperatureUseCase(&suite.CepValidator, &suite.TemperatureClient)
	output, err := useCase.Execute("99999999")

	// assert
	suite.NotNil(output)
	suite.NotNil(err)
	suite.Equal(GetTemperatureOutputDTO{}, output)
	suite.Equal("can not find zipcode", err.Error())
}
