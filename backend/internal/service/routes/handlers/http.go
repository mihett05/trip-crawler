package handlers

import (
	"github.com/mihett05/trip-crawler/internal/service/routes"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	logger           *zap.Logger
	itineraryBuilder routes.ItineraryBuilder
}

func NewHTTPHandler(logger *zap.Logger, builder routes.ItineraryBuilder) *HTTPHandler {
	return &HTTPHandler{
		logger:           logger,
		itineraryBuilder: builder,
	}
}
