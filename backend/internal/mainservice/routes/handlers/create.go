package handlers

import "net/http"

func (h *HTTPHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	h.logger.Warn("unimplemented")
	w.WriteHeader(http.StatusInternalServerError)
}
