package main

import (
	"log"
	"os"

	webSrv "github.com/nagahshi/pos_go_weather_otel/internal/infra/web/server"
	usecase "github.com/nagahshi/pos_go_weather_otel/internal/useCase"

	"github.com/nagahshi/pos_go_weather_otel/internal/infra/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("número de porta [PORT] não definido no ambiente")
		return
	}

	webserver := webSrv.NewWebServer(port)

	handler := web.NewHandler(
		*usecase.NewGetLatLonByCEPUseCase(),
		*usecase.NewGetWeatherUseCase(os.Getenv("WEATHER_API_KEY")),
	)

	webserver.AddHandler("POST", "/cep", handler.GetWeatherByCEP)
	webserver.AddHandler("GET", "/cep/{cep}", handler.GetWeather)

	webserver.Start()
}
