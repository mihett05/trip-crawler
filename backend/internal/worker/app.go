package worker

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/internal/worker/core/nats"
	queuehandlers "github.com/mihett05/trip-crawler/internal/worker/queue/handlers"
	"github.com/mihett05/trip-crawler/pkg/application"
	"go.uber.org/zap"
)

type App struct {
	App          *application.App
	NATS         *nats.Client
	QueueHandler *queuehandlers.Handler
}

func New(ctx context.Context, envFileName string) (*App, error) {
	app, err := application.New(ctx, envFileName)
	if err != nil {
		return nil, fmt.Errorf("application.New: %w", err)
	}

	natsClient, err := nats.New(ctx, app.Config, app.Observability.Logger)
	if err != nil {
		return nil, fmt.Errorf("nats.New: %w", err)
	}

	queueHandler := queuehandlers.New()

	return &App{
		App:          app,
		NATS:         natsClient,
		QueueHandler: queueHandler,
	}, nil
}

func (a *App) Start(ctx context.Context) {
	if err := a.NATS.RunConsumers(ctx, a.QueueHandler); err != nil {
		a.App.Observability.Logger.Fatal(
			"worker.App.NATS.RunConsumers: failed to start consumers",
			zap.Error(err),
		)
	}

	a.App.Observability.Logger.Info("worker started")
}

func (a *App) Shutdown() {
	a.NATS.Shutdown()
}
