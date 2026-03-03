package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mihett05/trip-crawler/pkg/application/config"
)

func initLoggerProvider(ctx context.Context, config config.Config, resource *resource.Resource) (
	*log.LoggerProvider,
	*zap.Logger,
	error,
) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(config.Observability.OTelExporter),
	}
	if config.Observability.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("otlploggrpc.New: %w", err)
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(resource),
	)
	otellogglobal.SetLoggerProvider(loggerProvider)

	baseZapLogger, err := getZapLogger(config.App.Environment)
	if err != nil {
		return nil, nil, fmt.Errorf("getZapLogger: %w", err)
	}

	zapOtelCore := otelzap.NewCore(config.App.Name, otelzap.WithLoggerProvider(loggerProvider))
	zapTee := zapcore.NewTee(baseZapLogger.Core(), zapOtelCore)

	return loggerProvider, zap.New(zapTee, zap.AddCaller()), nil
}

func getZapLogger(environment string) (*zap.Logger, error) {
	if environment == config.EnvDevelopment {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}
