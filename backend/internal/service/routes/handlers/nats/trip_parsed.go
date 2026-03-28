package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleTripParsed(ctx context.Context, request messages.TripParsed) error {
	// TODO: сохранение результата парсинга в dgraph
	return nil
}
