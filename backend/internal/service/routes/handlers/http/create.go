package handlers

import (
	"fmt"
	"net/http"

	"github.com/mihett05/trip-crawler/pkg/api"
)

func (h *Handler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	req, ok := api.ParseAndRespond[api.CreateRouteRequest](w, r)
	if !ok {
		return
	}

	minDays := req.DurationMinDays
	maxDays := minDays
	if req.DurationMaxDays != nil && *req.DurationMaxDays > minDays {
		maxDays = *req.DurationMaxDays
	}

	route, err := h.itineraryBuilder.Build(r.Context(), req.Points, req.StartDate.Time.Unix(), minDays, maxDays)
	if err != nil {
		api.Write(w, api.Error{
			Code:    api.INTERNALERROR,
			Message: fmt.Sprintf("failed to build itinerary: %v", err),
		}, http.StatusInternalServerError)
		return
	}

	if len(route) == 0 {
		api.Write(w, api.Error{
			Code:    "NOT_FOUND",
			Message: "no routes found for the specified criteria",
		}, http.StatusNotFound)
		return
	}

	// Преобразуем маршрут в API-формат
	apiPoints := make([]api.RoutePoint, 0, len(route))
	for _, point := range route {
		var detailsPtr *string
		if point.Details != "" {
			detailsPtr = &point.Details
		}

		var coordinates *api.Coordinates
		if point.Latitude != nil && point.Longitude != nil {
			coordinates = &api.Coordinates{
				Latitude:  *point.Latitude,
				Longitude: *point.Longitude,
			}
		}

		apiPoints = append(apiPoints, api.RoutePoint{
			Name:           point.Name,
			StartTimestamp: point.StartTimestamp,
			EndTimestamp:   point.EndTimestamp,
			Details:        detailsPtr,
			Coordinates:    coordinates,
		})
	}

	api.Write(w, api.CreateRouteResponse{Points: apiPoints}, http.StatusOK)
}
