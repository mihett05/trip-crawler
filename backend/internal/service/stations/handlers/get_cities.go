package handlers

import (
	"fmt"
	"net/http"

	"github.com/mihett05/trip-crawler/pkg/api"
)

func (h *HTTPHandler) GetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.service.GetCities(r.Context())
	if err != nil {
		api.Write(w, api.Error{
			Code:    api.INTERNALERROR,
			Message: fmt.Errorf("service.GetCities: %w", err).Error(),
		}, http.StatusInternalServerError)
		return
	}

	resp := api.GetCitiesResponse{
		Cities: &cities,
	}

	api.Write(w, resp, http.StatusOK)
}
