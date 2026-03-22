package handlers

import (
	"github.com/mihett05/trip-crawler/internal/service/stations/services/cities"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	logger  *zap.Logger
	service *cities.Service
}

func NewHTTPHandler(logger *zap.Logger, service *cities.Service) *HTTPHandler {
	return &HTTPHandler{
		logger:  logger,
		service: service,
	}
}
