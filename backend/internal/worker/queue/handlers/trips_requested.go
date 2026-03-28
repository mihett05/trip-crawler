package handlers

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleTripRequested(ctx context.Context, request messages.TripRequested) error {
	// TODO: связать с парсерами
	fmt.Println("trip requested", request)
	return nil
}
