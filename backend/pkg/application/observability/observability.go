package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.uber.org/zap"

	"github.com/mihett05/trip-crawler/pkg/application/config"
)

type Observability struct {
	Logger         *zap.Logger
	Resource       *resource.Resource
	MeterProvider  *metric.MeterProvider
	LoggerProvider *log.LoggerProvider
	TracerProvider *trace.TracerProvider
}

func New(ctx context.Context, cfg config.Config) (*Observability, error) {
	serviceResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.App.Name),
		semconv.DeploymentEnvironmentNameKey.String(cfg.App.Environment),
	)

	meterProvider, err := initMeterProvider(ctx, cfg, serviceResource)
	if err != nil {
		return nil, fmt.Errorf("initMeterProvider: %w", err)
	}

	loggerProvider, logger, err := initLoggerProvider(ctx, cfg, serviceResource)
	if err != nil {
		return nil, fmt.Errorf("initLoggerProvider: %w", err)
	}

	tracerProvider, err := initTracerProvider(ctx, cfg, serviceResource)
	if err != nil {
		return nil, fmt.Errorf("initTracerProvider: %w", err)
	}

	return &Observability{
		Resource:       serviceResource,
		MeterProvider:  meterProvider,
		LoggerProvider: loggerProvider,
		Logger:         logger,
		TracerProvider: tracerProvider,
	}, nil
}

func (o *Observability) Shutdown(ctx context.Context) error {
	if err := o.MeterProvider.Shutdown(ctx); err != nil {
		return err
	}

	if err := o.LoggerProvider.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
