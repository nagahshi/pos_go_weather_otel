package usecase

import (
	"context"

	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/nagahshi/pos_go_weather_otel/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GetLatLonByCEP struct{}

func NewGetLatLonByCEPUseCase() *GetLatLonByCEP {
	return &GetLatLonByCEP{}
}

// Execute - busca de latitude e longitude pelo CEP
func (c *GetLatLonByCEP) Execute(ctx context.Context, CEP string) (output dto.CEPOutput, err error) {
	tracer := otel.Tracer("useCase-GetLatLonByCEP-Execute")
	ctx, spanSearch := tracer.Start(ctx, "service_search_zipcode")
	defer spanSearch.End()

	spanSearch.AddEvent("zipcode to search", trace.WithAttributes(attribute.String("zipcode", CEP)))
	srvc := service.NewBrasilAPIService(CEP)

	spanSearch.AddEvent("try search")
	response, err := srvc.Search(ctx)
	if err != nil {
		spanSearch.AddEvent("error on search", trace.WithAttributes(attribute.String("error", err.Error())))
		return output, err
	}

	spanSearch.AddEvent("search success", trace.WithAttributes(attribute.String("latitude", response.Latitude), attribute.String("longitude", response.Longitude)))
	return response, err
}
