package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/valyala/fastjson"

	PKGHttpClient "github.com/nagahshi/pos_go_weather_otel/pkg/http"
)

type WeatherAPI struct {
	key        string
	Localidade string
}

func NewWeatherAPIService(key string, localidade string) *WeatherAPI {
	return &WeatherAPI{
		key:        key,
		Localidade: localidade,
	}
}

func (c *WeatherAPI) Search() (weatherAPIOutput dto.WeatherOutput, err error) {
	var client = PKGHttpClient.GetNewClient()

	if c.key == "" {
		return weatherAPIOutput, errors.New("chave de acesso não informada")
	}

	// realizo pesquisas cada um em sua rotina
	resp, err := client.Get("http://api.weatherapi.com/v1/current.json?key=" + c.key + "&q=" + c.Localidade)
	if err != nil {
		return weatherAPIOutput, errors.New("ocorreu um erro, ao buscar informações: " + err.Error())
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return weatherAPIOutput, errors.New("ocorreu um erro, ao ler informações")
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			return weatherAPIOutput, errors.New("ocorreu um erro, ao tratar informações")
		}

		weatherAPIOutput.C = v.Get("current").Get("temp_c").GetFloat64()
		weatherAPIOutput.F = (weatherAPIOutput.C * 1.8) + 32
		weatherAPIOutput.K = weatherAPIOutput.C + 273.15

		return weatherAPIOutput, nil
	}

	return weatherAPIOutput, fmt.Errorf("ocorreu um erro, ao buscar informações: %s status: %d", string(respBody), resp.StatusCode)
}
