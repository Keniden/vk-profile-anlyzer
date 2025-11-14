package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"

	"inteam/internal/config"
)

type Provider struct {
	TracerProvider *sdktrace.TracerProvider
}

func Init(ctx context.Context, cfg config.TelemetryConfig, logger *zap.Logger) (*Provider, error) {
	if !cfg.Enabled || cfg.OTLPEndpoint == "" {
		logger.Info("otel disabled")
		return &Provider{TracerProvider: trace.NewNoopTracerProvider().(*sdktrace.TracerProvider)}, nil
	}

	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(cfg.OTLPEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	logger.Info("otel tracer initialized", zap.String("endpoint", cfg.OTLPEndpoint))

	return &Provider{TracerProvider: tp}, nil
}

