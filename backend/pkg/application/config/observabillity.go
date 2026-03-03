package config

import "time"

type ObservabilityConfig struct {
	OTelExporter    string        `env:"OTEL_EXPORTER"`
	Insecure        bool          `env:"OTEL_INSECURE"`
	MetricsInterval time.Duration `env:"METRICS_INTERVAL" envDefault:"15s"`
}
