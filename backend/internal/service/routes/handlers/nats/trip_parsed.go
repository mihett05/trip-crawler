package nats

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleTripParsed(ctx context.Context, request messages.TripParsed) error {
	// TODO: сохранение результата парсинга в dgraph
	fmt.Println("trip parsed", request)
	return nil
}
