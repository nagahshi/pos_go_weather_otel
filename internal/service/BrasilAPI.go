package service

import (
	"context"
	"errors"
	"io"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/valyala/fastjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	PKGHttpClient "github.com/nagahshi/pos_go_weather_otel/pkg/http"
)

type BrasilAPI struct {
	CEP string
}

func NewBrasilAPIService(CEP string) *BrasilAPI {
	return &BrasilAPI{
		CEP: CEP,
	}
}

// Search - busca de clima pelo CEP
func (c *BrasilAPI) Search(ctx context.Context) (CEPOutput dto.CEPOutput, err error) {
	tracer := otel.Tracer("service-BrasilAPI-search")

	_, spanRequest := tracer.Start(ctx, "service_BrasilAPI_request")

	spanRequest.AddEvent("new client http")
	var client = PKGHttpClient.GetNewClient()

	spanRequest.AddEvent("zipcode to search", trace.WithAttributes(attribute.String("zipcode", c.CEP)))
	resp, err := client.Get("https://brasilapi.com.br/api/cep/v2/" + c.CEP)
	if err != nil {
		spanRequest.AddEvent("error on search", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequest.End()
		return CEPOutput, errors.New("ocorreu um erro, ao buscar informações")
	}

	spanRequest.AddEvent("read response")
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		spanRequest.AddEvent("error on read response", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequest.End()
		return CEPOutput, errors.New("ocorreu um erro, ao ler informações")
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		spanRequest.AddEvent("parse response")
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			spanRequest.AddEvent("error on parse response", trace.WithAttributes(attribute.String("error", err.Error())))
			spanRequest.End()
			return CEPOutput, errors.New("ocorreu um erro, ao tratar informações")
		}

		CEPOutput.Logradouro = string(v.GetStringBytes("street"))
		CEPOutput.Bairro = string(v.GetStringBytes("neighborhood"))
		CEPOutput.UF = string(v.GetStringBytes("state"))
		CEPOutput.CIDADE = string(v.GetStringBytes("city"))
		CEPOutput.Latitude = string(v.Get("location").Get("coordinates").GetStringBytes("latitude"))
		CEPOutput.Longitude = string(v.Get("location").Get("coordinates").GetStringBytes("longitude"))

		spanRequest.AddEvent(
			"response success",
			trace.WithAttributes(
				attribute.String("latitude", CEPOutput.Latitude),
				attribute.String("longitude", CEPOutput.Longitude),
			),
		)
		spanRequest.End()
		return CEPOutput, nil
	}

	spanRequest.AddEvent("response error", trace.WithAttributes(attribute.String("error", string(respBody))))
	spanRequest.End()
	return CEPOutput, errors.New("ocorreu um erro, ao buscar informações")
}
