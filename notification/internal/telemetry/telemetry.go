package telemetry

import (
	"context"
	"errors"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init configures OpenTelemetry tracing and metrics when an OTLP endpoint is configured.
// It returns a no-op shutdown function when tracing is disabled.
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) == "" {
		return func(context.Context) error { return nil }, nil
	}
	if configuredServiceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); configuredServiceName != "" {
		serviceName = configuredServiceName
	}
	if strings.TrimSpace(serviceName) == "" {
		serviceName = "notification-service"
	}

	serviceResource := resource.NewWithAttributes("", attribute.String("service.name", serviceName))
	resources, err := resource.Merge(resource.Default(), serviceResource)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	shutdowns := []func(context.Context) error{traceProvider.Shutdown}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_METRICS_EXPORTER")), "otlp") {
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			_ = traceProvider.Shutdown(ctx)
			return nil, err
		}
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
			sdkmetric.WithResource(resources),
		)
		otel.SetMeterProvider(meterProvider)
		shutdowns = append(shutdowns, meterProvider.Shutdown)
	}

	return func(ctx context.Context) error {
		errs := make([]error, 0, len(shutdowns))
		for _, shutdown := range shutdowns {
			errs = append(errs, shutdown(ctx))
		}
		return errors.Join(errs...)
	}, nil
}
