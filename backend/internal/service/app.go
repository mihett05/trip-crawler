package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	apphttp "github.com/mihett05/trip-crawler/internal/service/core/http"
	routeshandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers"
	"github.com/mihett05/trip-crawler/internal/service/routes/repositories/graph"
	"github.com/mihett05/trip-crawler/pkg/application"
)

type App struct {
	App       *application.App
	Server    *http.Server
	DGraph    *dgraph.Client
	GraphRepo *graph.Repository
}

func New(ctx context.Context, envFileName string) (*App, error) {
	app, err := application.New(ctx, envFileName)
	if err != nil {
		return nil, fmt.Errorf("application.New: %w", err)
	}

	dgraphClient, err := dgraph.New(app.Config)
	if err != nil {
		return nil, fmt.Errorf("dgraph.New: %w", err)
	}

	graphRepo := graph.NewRepository(dgraphClient)

	if err := graphRepo.InitSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize dgraph schema: %w", err)
	}

	routesHandler := routeshandlers.NewHTTPHandler(app.Observability.Logger)

	httpHandler := apphttp.NewHandler(app.Config, routesHandler)

	return &App{
		App:       app,
		DGraph:    dgraphClient,
		GraphRepo: graphRepo,
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", app.Config.HTTP.Port),
			Handler:      httpHandler,
			ReadTimeout:  app.Config.HTTP.ReadTimeout,
			WriteTimeout: app.Config.HTTP.WriteTimeout,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		},
	}, nil
}

func (a *App) Start() {
	go func() {
		a.App.Observability.Logger.Info("main service server started", zap.Uint16("port", a.App.Config.HTTP.Port))
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.App.Observability.Logger.Fatal("service.App.Server.ListenAndServe: server failed", zap.Error(err))
		}
	}()
}

func (a *App) Shutdown() {
	ctxShutdown, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer a.App.Shutdown(ctxShutdown)

	if err := a.Server.Shutdown(ctxShutdown); err != nil {
		a.App.Observability.Logger.Fatal("service.App.Server.Shutdown: server forced to shutdown", zap.Error(err))
	}
}
