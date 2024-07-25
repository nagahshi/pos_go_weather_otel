package dto

type CEPInput struct {
	CEP string
}

type CEPOutput struct {
	Logradouro string
	Bairro     string
	UF         string
	CIDADE     string
	Latitude   string
	Longitude  string
}

// ToWeatherInput converts a CEPOutput to a WeatherInput
func (c CEPOutput) ToWeatherInput() WeatherInput {
	var weatherInput WeatherInput

	weatherInput.Logradouro = c.Logradouro
	weatherInput.Bairro = c.Bairro
	weatherInput.UF = c.UF
	weatherInput.CIDADE = c.CIDADE
	weatherInput.Latitude = c.Latitude
	weatherInput.Longitude = c.Longitude

	return weatherInput
}
