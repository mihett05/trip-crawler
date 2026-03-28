package handlers

import (
	"github.com/mihett05/trip-crawler/internal/service/cities/services/cities"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service *cities.Service
}

func New(logger *zap.Logger, service *cities.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}
