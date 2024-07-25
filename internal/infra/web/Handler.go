package web

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/valyala/fastjson"

	useCase "github.com/nagahshi/pos_go_weather_otel/internal/useCase"

	PKGHttp "github.com/nagahshi/pos_go_weather_otel/pkg/http"
)

type Handler struct {
	GetLatLonByCEP       useCase.GetLatLonByCEP
	GetWeatherByLocation useCase.GetWeatherUseCase
}

func NewHandler(GetLatLonByCEP useCase.GetLatLonByCEP, GetWeatherByLocation useCase.GetWeatherUseCase) *Handler {
	return &Handler{
		GetLatLonByCEP:       GetLatLonByCEP,
		GetWeatherByLocation: GetWeatherByLocation,
	}
}

type DataRequest struct {
	CEP string `json:"cep"`
}

// GetWeatherByCEP - busca de clima pelo CEP
func (wh *Handler) GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	var re *regexp.Regexp = regexp.MustCompile("[0-9]+")

	data := DataRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, "cant decode zipcode", http.StatusUnprocessableEntity)
		return
	}

	CEP := data.CEP

	CEP = strings.Join(re.FindAllString(CEP, -1), "")
	if len(CEP) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	if os.Getenv("HOST_SERVICE_B") == "" {
		http.Error(w, "service B unavailable [HOST_SERVICE_B]", http.StatusUnprocessableEntity)
		return
	}

	// then request service B
	client := PKGHttp.GetNewClient()
	resp, err := client.Get(os.Getenv("HOST_SERVICE_B") + "/cep/" + CEP)
	if err != nil {
		http.Error(w, "cant get data", http.StatusUnprocessableEntity)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "cant read data service B", http.StatusUnprocessableEntity)
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			http.Error(w, "cant parse data service B", http.StatusUnprocessableEntity)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		response := make(map[string]interface{})
		response["city"] = string(v.GetStringBytes("city"))
		response["temp_C"] = v.GetFloat64("temp_C")
		response["temp_F"] = v.GetFloat64("temp_F")
		response["temp_K"] = v.GetFloat64("temp_K")

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		return
	}

	http.Error(w, string(respBody), resp.StatusCode)
}

// GetWeather - busca de clima pelo CEP
func (wh *Handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	var CEP = chi.URLParam(r, "cep")

	// SearchCEP - busca de endereço pelo CEP
	outputCEP, err := wh.GetLatLonByCEP.Execute(CEP)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	input := outputCEP.ToWeatherInput()
	// GetWeather - busca de clima pelo endereço encontrado em SearchCEP
	outputWeather, err := wh.GetWeatherByLocation.Execute(input)
	if err != nil {
		http.Error(w, "can not find location to weather", http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	// hidratando com cidade
	outputWeather.City = outputCEP.CIDADE
	err = json.NewEncoder(w).Encode(outputWeather)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}
