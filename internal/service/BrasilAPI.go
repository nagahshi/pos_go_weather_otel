package service

import (
	"errors"
	"io"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/valyala/fastjson"

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

func (c *BrasilAPI) Search() (CEPOutput dto.CEPOutput, err error) {
	var client = PKGHttpClient.GetNewClient()

	resp, err := client.Get("https://brasilapi.com.br/api/cep/v2/" + c.CEP)
	if err != nil {
		return CEPOutput, errors.New("ocorreu um erro, ao buscar informações")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CEPOutput, errors.New("ocorreu um erro, ao ler informações")
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			return CEPOutput, errors.New("ocorreu um erro, ao tratar informações")
		}

		CEPOutput.Logradouro = string(v.GetStringBytes("street"))
		CEPOutput.Bairro = string(v.GetStringBytes("neighborhood"))
		CEPOutput.UF = string(v.GetStringBytes("state"))
		CEPOutput.CIDADE = string(v.GetStringBytes("city"))
		CEPOutput.Latitude = string(v.Get("location").Get("coordinates").GetStringBytes("latitude"))
		CEPOutput.Longitude = string(v.Get("location").Get("coordinates").GetStringBytes("longitude"))

		return CEPOutput, nil
	}

	return CEPOutput, errors.New("ocorreu um erro, ao buscar informações")
}
