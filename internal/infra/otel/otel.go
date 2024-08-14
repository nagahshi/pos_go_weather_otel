package otel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// SetupOTelSDK - configures the OpenTelemetry SDK with the given service name.
func SetupOTelSDK(serviceName string, ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	otel.SetTextMapPropagator(prop)

	tracerProvider, err := newTraceProvider(ctx, serviceName)
	if err != nil {
		return shutdown, errors.Join(err, shutdown(ctx))
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return
}

// newTraceProvider - creates a new trace provider configured with the given service name.
// func newTraceProvider(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
// 	zipkinEndpoint := os.Getenv("ZIPKIN_ENDPOINT")
// 	if zipkinEndpoint == "" {
// 		return nil, errors.New("zipkin [ZIPKIN_ENDPOINT] not configured yet")
// 	}

// 	traceExporter, err := zipkin.New(zipkinEndpoint)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
// 	if err != nil {
// 		return nil, err
// 	}

// 	bsp := trace.NewBatchSpanProcessor(traceExporter)
// 	traceProvider := trace.NewTracerProvider(
// 		trace.WithSampler(trace.AlwaysSample()),
// 		trace.WithResource(res),
// 		trace.WithSpanProcessor(bsp),
// 	)

// 	return traceProvider, nil
// }

func newTraceProvider(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
	collectorEndpoint := os.Getenv("COLLECTOR_ENDPOINT")
	if collectorEndpoint == "" {
		return nil, errors.New("[COLLECTOR_ENDPOINT] not configured yet")
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(collectorEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %v", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(serviceName),
	))
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}
