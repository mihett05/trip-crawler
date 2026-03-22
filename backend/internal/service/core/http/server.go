package http

import (
	"net/http"

	routes "github.com/mihett05/trip-crawler/internal/service/routes/handlers"
)

type Server struct {
	routes *routes.HTTPHandler
}

func NewServer(routes *routes.HTTPHandler) *Server {
	return &Server{
		routes: routes,
	}
}

func (s *Server) CreateRoute(w http.ResponseWriter, r *http.Request) {
	s.routes.CreateRoute(w, r)
}
