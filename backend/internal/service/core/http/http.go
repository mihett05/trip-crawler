package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	routeshandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers"
	"github.com/mihett05/trip-crawler/pkg/application/config"
)

func NewHandler(config config.Config, routes *routeshandlers.HTTPHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)

	r.Use(
		otelhttp.NewMiddleware(
			config.App.Name,
			otelhttp.WithServerName(config.App.Name),
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		),
	)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/routes", func(r chi.Router) {
			r.Post("/", routes.CreateRoute)
		})
	})

	return r
}
