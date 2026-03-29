package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	citiesqueue "github.com/mihett05/trip-crawler/internal/service/cities/gateways/queue"
	citieshttphandlers "github.com/mihett05/trip-crawler/internal/service/cities/handlers/http"
	citiesnatshandlers "github.com/mihett05/trip-crawler/internal/service/cities/handlers/nats"
	citiessservices "github.com/mihett05/trip-crawler/internal/service/cities/services/cities"
	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	apphttp "github.com/mihett05/trip-crawler/internal/service/core/http"
	"github.com/mihett05/trip-crawler/internal/service/core/nats"
	"github.com/mihett05/trip-crawler/internal/service/routes"
	"github.com/mihett05/trip-crawler/internal/service/routes/gateway"
	routesqueue "github.com/mihett05/trip-crawler/internal/service/routes/gateway/queue"
	routeshttphandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers/http"
	routesnatshandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers/nats"
	"github.com/mihett05/trip-crawler/internal/service/routes/repositories/graph"
	"github.com/mihett05/trip-crawler/internal/service/scheduler"
	"github.com/mihett05/trip-crawler/pkg/application"
)

type App struct {
	App              *application.App
	Server           *http.Server
	DGraph           *dgraph.Client
	NATS             *nats.Client
	GraphRepo        *graph.Repository
	ItineraryBuilder routes.ItineraryBuilder

	CitiesQueue *citiesqueue.Gateway
	RoutesQueue *routesqueue.Gateway

	CitiesNATSHandler *citiesnatshandlers.Handler
	RoutesNATSHandler *routesnatshandlers.Handler

	RoutesService *routes.Service

	Scheduler *scheduler.Scheduler
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

	citiesService := citiessservices.New(graphRepo)
	routesService := routes.New(graphRepo)

	natsClient, err := nats.New(ctx, app.Config, app.Observability.Logger)
	if err != nil {
		return nil, fmt.Errorf("nats.New: %w", err)
	}

	itineraryBuilder := gateway.NewDgraphItineraryBuilder(dgraphClient)

	citiesQueue := citiesqueue.New(natsClient)
	routesQueue := routesqueue.New(natsClient)

	scheduler := scheduler.New(app.Config.Scheduler)

	routesHTTPHandler := routeshttphandlers.New(app.Observability.Logger, itineraryBuilder)
	routesNATSHandler := routesnatshandlers.New(routesService)

	citiesHTTPHandler := citieshttphandlers.New(app.Observability.Logger, citiesService)
	citiesNATSHandler := citiesnatshandlers.New(citiesService)

	httpHandler := apphttp.NewHandler(app.Config, app.Observability.Logger, routesHTTPHandler, citiesHTTPHandler)

	return &App{
		App:               app,
		DGraph:            dgraphClient,
		NATS:              natsClient,
		GraphRepo:         graphRepo,
		ItineraryBuilder:  itineraryBuilder,
		CitiesQueue:       citiesQueue,
		RoutesQueue:       routesQueue,
		CitiesNATSHandler: citiesNATSHandler,
		RoutesNATSHandler: routesNATSHandler,
		RoutesService:     routesService,
		Scheduler:         scheduler,
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

func (a *App) Start(ctx context.Context) {
	go func() {
		a.App.Observability.Logger.Info("main service server started", zap.Uint16("port", a.App.Config.HTTP.Port))
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.App.Observability.Logger.Fatal("service.App.Server.ListenAndServe: server failed", zap.Error(err))
		}
	}()

	if err := a.NATS.RunConsumers(ctx, a.CitiesNATSHandler, a.RoutesNATSHandler); err != nil {
		a.App.Observability.Logger.Fatal(
			"worker.App.NATS.RunConsumers: failed to start consumers",
			zap.Error(err),
		)
	}
}

func (a *App) Shutdown() {
	ctxShutdown, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer a.App.Shutdown(ctxShutdown)

	if err := a.Server.Shutdown(ctxShutdown); err != nil {
		a.App.Observability.Logger.Fatal("service.App.Server.Shutdown: server forced to shutdown", zap.Error(err))
	}

	a.NATS.Shutdown()
}
