package viacep

type TemperatureInfo struct {
	City  string  `json:"city"`   // Nome da Cidade
	TempC float64 `json:"temp_C"` // Celsius
	TempF float64 `json:"temp_F"` // Fahrenheit
	TempK float64 `json:"temp_K"` // Kelvin
}
