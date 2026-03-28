package handlers

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleCitiesRequested(ctx context.Context, request messages.CitiesRequested) error {
	// TODO: связать с парсером
	fmt.Println("cities requested", request)
	return nil
}
