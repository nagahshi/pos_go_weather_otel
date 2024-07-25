package usecase

import (
	"github.com/nagahshi/pos_go_weather_otel/internal/dto"
	"github.com/nagahshi/pos_go_weather_otel/internal/service"
)

type GetLatLonByCEP struct{}

func NewGetLatLonByCEPUseCase() *GetLatLonByCEP {
	return &GetLatLonByCEP{}
}

func (c *GetLatLonByCEP) Execute(CEP string) (output dto.CEPOutput, err error) {
	srvc := service.NewBrasilAPIService(CEP)

	response, err := srvc.Search()
	if err != nil {
		return output, err
	}

	return response, err
}
