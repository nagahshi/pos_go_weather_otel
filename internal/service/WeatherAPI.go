package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/valyala/fastjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

// Search - busca de clima pelo local
func (c *WeatherAPI) Search(ctx context.Context) (weatherAPIOutput dto.WeatherOutput, err error) {
	tracer := otel.Tracer("service-weatherAPI-search")

	_, spanRequest := tracer.Start(ctx, "service_weatherAPI_request")

	spanRequest.AddEvent("new client http")
	var client = PKGHttpClient.GetNewClient()

	if c.key == "" {
		spanRequest.AddEvent("key[WEATHER_API_KEY] not found")
		spanRequest.End()
		return weatherAPIOutput, errors.New("chave de acesso não informada")
	}

	spanRequest.AddEvent("localidade to search", trace.WithAttributes(attribute.String("localidade", c.Localidade)))
	// realizo pesquisas cada um em sua rotina
	resp, err := client.Get("http://api.weatherapi.com/v1/current.json?key=" + c.key + "&q=" + c.Localidade)
	if err != nil {
		spanRequest.AddEvent("error on search", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequest.End()
		return weatherAPIOutput, errors.New("ocorreu um erro, ao buscar informações: " + err.Error())
	}

	spanRequest.AddEvent("read response")
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		spanRequest.AddEvent("error on read response", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequest.End()
		return weatherAPIOutput, errors.New("ocorreu um erro, ao ler informações")
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		spanRequest.AddEvent("parse response")
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			spanRequest.AddEvent("error on parse response", trace.WithAttributes(attribute.String("error", err.Error())))
			spanRequest.End()
			return weatherAPIOutput, errors.New("ocorreu um erro, ao tratar informações")
		}

		weatherAPIOutput.C = v.Get("current").Get("temp_c").GetFloat64()
		weatherAPIOutput.F = (weatherAPIOutput.C * 1.8) + 32
		weatherAPIOutput.K = weatherAPIOutput.C + 273.15

		spanRequest.AddEvent(
			"response success",
			trace.WithAttributes(
				attribute.Float64("temp_C", weatherAPIOutput.C),
				attribute.Float64("temp_F", weatherAPIOutput.F),
				attribute.Float64("temp_K", weatherAPIOutput.K),
			),
		)
		spanRequest.End()

		return weatherAPIOutput, nil
	}

	spanRequest.AddEvent(
		"ocorreu um erro",
		trace.WithAttributes(
			attribute.String("erro", fmt.Sprintf("ocorreu um erro, ao buscar informações: %s status: %d", string(respBody), resp.StatusCode)),
		),
	)
	spanRequest.End()

	return weatherAPIOutput, fmt.Errorf("ocorreu um erro, ao buscar informações: %s status: %d", string(respBody), resp.StatusCode)
}
