package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleTripParsed(ctx context.Context, request messages.TripParsed) error {
	t := request.Trip

	tripModel := &models.Trip{
		ExternalID:  t.ExternalID,
		DepartureAt: time.Unix(t.DepartureAtTimestamp, 0),
		ArrivalAt:   time.Unix(t.ArrivalAtTimestamp, 0),
		Price:       0,
	}

	if len(t.Availability) > 0 {
		tripModel.Price = t.Availability[0].Price
	}

	if t.DestinationStationID != "" {
		tripModel.Destination = &models.Station{ID: t.DestinationStationID}
	}

	if err := h.repo.SaveTrip(ctx, tripModel); err != nil {
		return fmt.Errorf("failed to save trip %s: %w", t.ExternalID, err)
	}

	return nil
}
