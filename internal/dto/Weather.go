package dto

type WeatherInput struct {
	Logradouro string
	Bairro     string
	UF         string
	CIDADE     string
	Latitude   string
	Longitude  string
}

type WeatherOutput struct {
	City string  `json:"city"`
	C    float64 `json:"temp_C"`
	F    float64 `json:"temp_F"`
	K    float64 `json:"temp_K"`
}
