package mainservice

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	apphttp "github.com/mihett05/trip-crawler/internal/mainservice/http"
	routeshandlers "github.com/mihett05/trip-crawler/internal/mainservice/routes/handlers"
	"github.com/mihett05/trip-crawler/pkg/application"
	"go.uber.org/zap"
)

const envFileName = ".env.mainservice.local"

type App struct {
	App    *application.App
	Server *http.Server
}

func New(ctx context.Context) (*App, error) {
	app, err := application.New(ctx, envFileName)
	if err != nil {
		return nil, fmt.Errorf("application.New: %w", err)
	}

	routesHandler := routeshandlers.NewHTTPHandler(app.Observability.Logger)

	httpHandler := apphttp.NewHandler(app.Config, routesHandler)

	return &App{
		App: app,
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

func (a *App) Run() {
	go func() {
		a.App.Observability.Logger.Info("main service server started", zap.Uint16("port", a.App.Config.HTTP.Port))
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.App.Observability.Logger.Fatal("mainservice.App.Server.ListenAndServe: server failed", zap.Error(err))
		}
	}()

	a.App.WaitUntilStop()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer a.App.Shutdown(ctxShutdown)

	if err := a.Server.Shutdown(ctxShutdown); err != nil {
		a.App.Observability.Logger.Fatal("mainservice.App.Server.Shutdown: server forced to shutdown", zap.Error(err))
	}
}
