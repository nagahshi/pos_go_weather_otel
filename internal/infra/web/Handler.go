package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/traceid"
	"github.com/go-chi/transport"
	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/nagahshi/pos_go_weather_otel/internal/usecase"
	"github.com/valyala/fastjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	GetLatLonByCEP       usecase.GetLatLonByCEP
	GetWeatherByLocation usecase.GetWeatherUseCase
}

// NewHandler - cria um novo handler com os usecases
func NewHandler(GetLatLonByCEP usecase.GetLatLonByCEP, GetWeatherByLocation usecase.GetWeatherUseCase) *Handler {
	return &Handler{
		GetLatLonByCEP:       GetLatLonByCEP,
		GetWeatherByLocation: GetWeatherByLocation,
	}
}

// GetLocationByCEPRequest - estrutura de entrada para busca de clima pelo CEP
type GetLocationByCEPRequest struct {
	CEP string `json:"cep"`
}

// GetWeatherByLocalRequest - estrutura de entrada para busca de clima pelo local
type GetWeatherByLocalRequest struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// GetLocationByCEP - busca de clima pelo CEP
func (wh *Handler) GetLocationByCEP(w http.ResponseWriter, r *http.Request) {
	var re *regexp.Regexp = regexp.MustCompile("[0-9]+")

	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := otel.Tracer("handler-GetLocationByCEP")
	ctx, spanValidate := tracer.Start(ctx, "validate_zipcode")

	spanValidate.AddEvent("extract POST body")
	data := GetLocationByCEPRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		spanValidate.AddEvent("error on decode body", trace.WithAttributes(attribute.String("error", err.Error())))
		spanValidate.End()
		http.Error(w, "cant decode zipcode", http.StatusUnprocessableEntity)
		return
	}

	spanValidate.AddEvent("sanitize zipcode", trace.WithAttributes(attribute.String("zipcode", data.CEP)))
	CEP := strings.Join(re.FindAllString(data.CEP, -1), "")
	if len(CEP) != 8 {
		spanValidate.AddEvent("error on check validate zipcode", trace.WithAttributes(attribute.String("error", err.Error())))
		spanValidate.End()
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	spanValidate.AddEvent("sanitized zipcode", trace.WithAttributes(attribute.String("zipcode", CEP)))
	spanValidate.End()

	ctx, spanSearch := tracer.Start(ctx, "zipcode-search")
	spanSearch.AddEvent("search location by zipcode")
	outputCEP, err := wh.GetLatLonByCEP.Execute(ctx, CEP)
	if err != nil {
		spanSearch.AddEvent("error on search location", trace.WithAttributes(attribute.String("error", err.Error())))
		spanSearch.End()
		http.Error(w, "can not find location to weather", http.StatusNotFound)
		return
	}

	spanSearch.AddEvent("prepare request to service B")
	locationJson, err := json.Marshal(outputCEP)
	if err != nil {
		spanSearch.AddEvent("error on prepare request", trace.WithAttributes(attribute.String("error", err.Error())))
		spanSearch.End()
		http.Error(w, "can not find location to weather", http.StatusNotFound)
		return
	}

	spanSearch.End()

	ctx, spanRequestServiceB := tracer.Start(ctx, "CEP-request-service-B")

	spanRequestServiceB.AddEvent("prepare transport")
	http.DefaultTransport = transport.Chain(
		http.DefaultTransport,
		transport.SetHeader("User-Agent", "my-app/v1.0.0"),
		traceid.Transport,
	)

	spanRequestServiceB.AddEvent("prepare request")
	req, err := http.NewRequest("POST", os.Getenv("HOST_SERVICE_B")+"/weather", strings.NewReader(string(locationJson)))
	if err != nil {
		spanRequestServiceB.AddEvent("request error service B", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequestServiceB.End()
		http.Error(w, "cant create request", http.StatusUnprocessableEntity)
		return
	}

	spanRequestServiceB.AddEvent("try request service B")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		spanRequestServiceB.AddEvent("request error service B", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequestServiceB.End()
		http.Error(w, "cant get data", http.StatusUnprocessableEntity)
		return
	}

	spanRequestServiceB.AddEvent("read data response service B")
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		spanRequestServiceB.AddEvent("error read body service B", trace.WithAttributes(attribute.String("error", err.Error())))
		spanRequestServiceB.End()
		http.Error(w, "cant read data service B", http.StatusUnprocessableEntity)
		return
	}
	spanRequestServiceB.End()

	_, spanResponse := tracer.Start(ctx, "CEP-response")
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		spanResponse.AddEvent("response service B success")
		var p fastjson.Parser
		v, err := p.Parse(string(respBody))
		if err != nil {
			spanResponse.AddEvent("error parse body data service B", trace.WithAttributes(attribute.String("error", err.Error())))
			http.Error(w, "cant parse data service B", http.StatusUnprocessableEntity)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		spanResponse.AddEvent("prepare to response")

		response := make(map[string]interface{})
		response["city"] = outputCEP.CIDADE
		response["temp_C"] = v.GetFloat64("temp_C")
		response["temp_F"] = v.GetFloat64("temp_F")
		response["temp_K"] = v.GetFloat64("temp_K")

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			spanResponse.AddEvent("error response service B", trace.WithAttributes(attribute.String("error", err.Error())))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		spanResponse.AddEvent(
			"response success",
			trace.WithAttributes(
				attribute.String("city", outputCEP.CIDADE),
				attribute.Float64("temp_C", v.GetFloat64("temp_C")),
				attribute.Float64("temp_F", v.GetFloat64("temp_F")),
				attribute.Float64("temp_K", v.GetFloat64("temp_K")),
			),
		)
		spanResponse.End()

		return
	}

	spanResponse.AddEvent(fmt.Sprintf("response service B error: %d", resp.StatusCode))
	spanResponse.End()
	http.Error(w, string(respBody), resp.StatusCode)
}

// GetWeatherByLocal - busca de clima pelo local
func (wh *Handler) GetWeatherByLocal(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := traceid.NewContext(r.Context())
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := otel.Tracer("handler-GetWeatherByLocal")
	ctx, spanValidate := tracer.Start(ctx, "validate_location")

	spanValidate.AddEvent("extract POST body")
	data := GetWeatherByLocalRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		spanValidate.AddEvent("error on decode body", trace.WithAttributes(attribute.String("error", err.Error())))
		spanValidate.End()
		http.Error(w, "cant decode location", http.StatusUnprocessableEntity)
		return
	}
	spanValidate.End()

	ctx, spanInput := tracer.Start(ctx, "weather_input")
	input := dto.WeatherInput{
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
	}

	spanInput.AddEvent("location data", trace.WithAttributes(attribute.String("latitude", data.Latitude), attribute.String("longitude", data.Longitude)))

	spanInput.End()

	ctx, spanSearch := tracer.Start(ctx, "weather_search")
	outputWeather, err := wh.GetWeatherByLocation.Execute(ctx, input)
	if err != nil {
		spanSearch.AddEvent("error on search weather: cant find location")
		spanSearch.End()
		http.Error(w, "can not find location to weather", http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	spanSearch.AddEvent("locations and weather found")
	spanSearch.End()

	_, spanResponse := tracer.Start(ctx, "weather_response")
	spanResponse.AddEvent("prepare response")
	// hidratando com cidade
	err = json.NewEncoder(w).Encode(outputWeather)
	if err != nil {
		spanResponse.AddEvent("error on response", trace.WithAttributes(attribute.String("error", err.Error())))
		spanResponse.End()
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	spanResponse.AddEvent(
		"response success",
		trace.WithAttributes(
			attribute.Float64("celcius", outputWeather.C),
			attribute.Float64("fahrenheit", outputWeather.F),
			attribute.Float64("kelvin", outputWeather.K),
		),
	)
	spanResponse.End()
}
