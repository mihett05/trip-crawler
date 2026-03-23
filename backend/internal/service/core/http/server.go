package http

import (
	"net/http"

	routes "github.com/mihett05/trip-crawler/internal/service/routes/handlers"
	stations "github.com/mihett05/trip-crawler/internal/service/stations/handlers"
)

type Server struct {
	routes   *routes.HTTPHandler
	stations *stations.HTTPHandler
}

func NewServer(routes *routes.HTTPHandler, stations *stations.HTTPHandler) *Server {
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
