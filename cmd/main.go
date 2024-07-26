package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/nagahshi/pos_go_weather_otel/internal/infra/otel"
	"github.com/nagahshi/pos_go_weather_otel/internal/infra/web"
	"github.com/nagahshi/pos_go_weather_otel/internal/usecase"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("server port [PORT] not configured yet")
		return
	}
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		log.Fatal("service name [SERVICE_NAME] not configured yet")
		return
	}

	handler := web.NewHandler(
		*usecase.NewGetLatLonByCEPUseCase(),
		*usecase.NewGetWeatherUseCase(os.Getenv("WEATHER_API_KEY")),
	)

	ctx := context.Background()
	// Setup OTel SDK
	otelShutdown, err := otel.SetupOTelSDK(serviceName, ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer otelShutdown(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/cep", handler.GetLocationByCEP)
	mux.HandleFunc("/weather", handler.GetWeatherByLocal)

	srv := &http.Server{
		Addr:         ":" + port,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      otelhttp.NewHandler(mux, "/"),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
