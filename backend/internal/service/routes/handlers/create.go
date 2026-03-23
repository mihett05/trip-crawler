package handlers

import (
	"fmt"
	"net/http"

	"github.com/mihett05/trip-crawler/pkg/api"
)

func (h *HTTPHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	req, ok := api.ParseAndRespond[api.CreateRouteRequest](w, r)
	if !ok {
		return
	}

	points, err := h.itineraryBuilder.Build(r.Context(), req.Points, req.StartDate.Unix())
	if err != nil {
		api.Write(w, api.Error{
			Code:    api.INTERNALERROR,
			Message: fmt.Errorf("itineraryBuilder.Build: %w", err).Error(),
		}, http.StatusInternalServerError)
		return
	}

	apiPoints := make([]api.RoutePoint, 0, len(points))
	for _, point := range points {
		apiPoints = append(apiPoints, api.RoutePoint{
			Coordinates:    nil,
			Details:        &point.Details,
			EndTimestamp:   point.EndTimestamp,
			Name:           point.Name,
			StartTimestamp: point.StartTimestamp,
		})
	}

	resp := api.CreateRouteResponse{
		Points: apiPoints,
	}

	api.Write(w, resp, http.StatusOK)
}
