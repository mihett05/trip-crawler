package application

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/application/config"
	"github.com/mihett05/trip-crawler/pkg/application/observability"
)

type App struct {
	Config        config.Config
	Observability *observability.Observability
}

func New(ctx context.Context, envFileName string) (*App, error) {
	cfg, err := config.New(envFileName)
	if err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	obs, err := observability.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("observability.New: %w", err)
	}

	return &App{
		Config:        cfg,
		Observability: obs,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) {
	defer a.Observability.Shutdown(ctx)
}
