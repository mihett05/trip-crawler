package http

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	routehandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers"
	stationhandlers "github.com/mihett05/trip-crawler/internal/service/stations/handlers"
	"github.com/mihett05/trip-crawler/pkg/api"
	"github.com/mihett05/trip-crawler/pkg/application/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

func NewHandler(config config.Config, logger *zap.Logger, routes *routehandlers.HTTPHandler, stations *stationhandlers.HTTPHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(ZapLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{}))

	r.Use(
		otelhttp.NewMiddleware(
			config.App.Name,
			otelhttp.WithServerName(config.App.Name),
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		),
	)

	server := NewServer(routes, stations)
	api.HandlerFromMux(server, r)

	r.Get("/openapi.json", ServeOpenAPI)
	r.Get("/docs", ServeSwagger)

	return r
}
