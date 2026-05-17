package worker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mihett05/trip-crawler/internal/worker/core/nats"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	workergateway "github.com/mihett05/trip-crawler/internal/worker/queue/gateways"
	queuehandlers "github.com/mihett05/trip-crawler/internal/worker/queue/handlers"
	"github.com/mihett05/trip-crawler/pkg/application"
	"go.uber.org/zap"
)

type App struct {
	App          *application.App
	Server       *http.Server
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

	rzdClient := rzd.NewClient(30 * time.Second)
	gateway := workergateway.New(natsClient.Connection.JetStream)
	queueHandler := queuehandlers.New(rzdClient, gateway)

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	return &App{
		App:          app,
		NATS:         natsClient,
		QueueHandler: queueHandler,
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", app.Config.HTTP.Port),
			Handler:      r,
			ReadTimeout:  app.Config.HTTP.ReadTimeout,
			WriteTimeout: app.Config.HTTP.WriteTimeout,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		},
	}, nil
}

func (a *App) Start(ctx context.Context) {
	go func() {
		a.App.Observability.Logger.Info("worker server started", zap.Uint16("port", a.App.Config.HTTP.Port))
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.App.Observability.Logger.Fatal("worker.App.Server.ListenAndServe: server failed", zap.Error(err))
		}
	}()
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
