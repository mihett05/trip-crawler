package handlers

import (
	"github.com/mihett05/trip-crawler/internal/service/routes"
	"go.uber.org/zap"
)

type Handler struct {
	logger           *zap.Logger
	itineraryBuilder routes.ItineraryBuilder
}

func New(logger *zap.Logger, builder routes.ItineraryBuilder) *Handler {
	return &Handler{
		logger:           logger,
		itineraryBuilder: builder,
	}
}
