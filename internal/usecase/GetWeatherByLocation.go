package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/nagahshi/pos_go_weather_otel/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GetWeatherUseCase struct {
	key string
}

func NewGetWeatherUseCase(key string) *GetWeatherUseCase {
	return &GetWeatherUseCase{
		key: key,
	}
}

// Execute - busca de clima pelo local
func (c *GetWeatherUseCase) Execute(ctx context.Context, weatherInput dto.WeatherInput) (output dto.WeatherOutput, err error) {
	tracer := otel.Tracer("useCase-GetWeather-Execute")
	ctx, spanSearch := tracer.Start(ctx, "service_search_weather")
	defer spanSearch.End()

	if c.key == "" {
		spanSearch.AddEvent("key[WEATHER_API_KEY] not found")
		return output, errors.New("chave de consulta [WEATHER_API_KEY] não encontrada")
	}

	spanSearch.AddEvent(
		"weather input",
		trace.WithAttributes(
			attribute.String("latitude", weatherInput.Latitude),
			attribute.String("longitude", weatherInput.Longitude),
			attribute.String("cidade", weatherInput.CIDADE),
			attribute.String("uf", weatherInput.UF),
		),
	)

	// se houver latitude e longitude, utilizo elas
	local := weatherInput.Latitude + "," + weatherInput.Longitude
	if weatherInput.Latitude == "" || weatherInput.Longitude == "" {
		// se não houver latitude e longitude, utilizo cidade e estado
		local = strings.ToLower(weatherInput.CIDADE + "," + weatherInput.UF)
	}

	spanSearch.AddEvent("try search")
	srvc := service.NewWeatherAPIService(c.key, local)
	responseWeatherAPI, err := srvc.Search(ctx)
	if err != nil {
		spanSearch.AddEvent("error on search", trace.WithAttributes(attribute.String("error", err.Error())))
		return output, err
	}

	spanSearch.AddEvent(
		"search success",
		trace.WithAttributes(
			attribute.Float64("temp_C", responseWeatherAPI.C),
			attribute.Float64("temp_F", responseWeatherAPI.F),
			attribute.Float64("temp_K", responseWeatherAPI.K),
		),
	)

	return responseWeatherAPI, err
}
