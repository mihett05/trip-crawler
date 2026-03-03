package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App           AppConfig           `envPrefix:"APP_"`
	HTTP          HTTPConfig          `envPrefix:"HTTP_"`
	Observability ObservabilityConfig `envPrefix:"OBSERVABILITY_"`
}

func New(filename string) (Config, error) {
	err := godotenv.Load(filename)
	if err != nil {
		fmt.Println("local env file is skipped", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("env.ParseAs: %w", err)
	}

	return cfg, nil
}
