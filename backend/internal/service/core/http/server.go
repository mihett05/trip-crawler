package http

import (
	"net/http"

	cities "github.com/mihett05/trip-crawler/internal/service/cities/handlers/http"
	routes "github.com/mihett05/trip-crawler/internal/service/routes/handlers/http"
)

type Server struct {
	routes   *routes.Handler
	stations *cities.Handler
}

func NewServer(routes *routes.Handler, stations *cities.Handler) *Server {
	return &Server{
		routes:   routes,
		stations: stations,
	}
}

func (s *Server) CreateRoute(w http.ResponseWriter, r *http.Request) {
	s.routes.CreateRoute(w, r)
}

func (s *Server) GetCities(w http.ResponseWriter, r *http.Request) {
	s.stations.GetCities(w, r)
}
