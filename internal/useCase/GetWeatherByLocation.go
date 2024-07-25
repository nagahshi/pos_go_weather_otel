package usecase

import (
	"errors"
	"strings"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/nagahshi/pos_go_weather_otel/internal/service"
)

type GetWeatherUseCase struct {
	key string
}

func NewGetWeatherUseCase(key string) *GetWeatherUseCase {
	return &GetWeatherUseCase{
		key: key,
	}
}

func (c *GetWeatherUseCase) Execute(weatherInput dto.WeatherInput) (output dto.WeatherOutput, err error) {
	if c.key == "" {
		return output, errors.New("chave de consulta [WEATHER_API_KEY] não encontrada")
	}

	// se houver latitude e longitude, utilizo elas
	local := weatherInput.Latitude + "," + weatherInput.Longitude
	if weatherInput.Latitude == "" || weatherInput.Longitude == "" {
		// se não houver latitude e longitude, utilizo cidade e estado
		local = strings.ToLower(weatherInput.CIDADE + "," + weatherInput.UF)
	}

	srvc := service.NewWeatherAPIService(c.key, local)
	responseWeatherAPI, err := srvc.Search()
	if err != nil {
		return output, err
	}

	return responseWeatherAPI, err
}
