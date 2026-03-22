package handlers

import (
	"net/http"

	"github.com/mihett05/trip-crawler/pkg/api"
)

func (h *HTTPHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	_, ok := api.ParseAndRespond[api.CreateRouteRequest](w, r)
	if !ok {
		return
	}

	resp := api.CreateRouteResponse{
		Points: []api.RoutePoint{},
	}

	api.Write(w, resp, http.StatusOK)
}
