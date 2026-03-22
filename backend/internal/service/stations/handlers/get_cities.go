package handlers

import (
	"net/http"

	"github.com/mihett05/trip-crawler/pkg/api"
)

func (h *HTTPHandler) GetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.service.GetCities(r.Context())
	if err != nil {
		api.Write(w, api.Error{
			Code:    api.INTERNALERROR,
			Message: err.Error(),
		}, http.StatusInternalServerError)
	}

	resp := api.GetCitiesResponse{
		Cities: &cities,
	}

	api.Write(w, resp, http.StatusOK)
}
