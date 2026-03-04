package handlers

import (
	"go.uber.org/zap"
)

type HTTPHandler struct {
	logger *zap.Logger
}

func NewHTTPHandler(logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		logger: logger,
	}
}
