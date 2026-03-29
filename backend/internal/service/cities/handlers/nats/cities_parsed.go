package nats

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleCitiesParsed(ctx context.Context, request messages.CitiesParsed) error {
	for _, c := range request.Cities {
		cityModel := &models.City{
			Name:      c.Name,
			Latitude:  c.Coordinates.Latitude,
			Longitude: c.Coordinates.Longitude,
		}

		for _, s := range c.Stations {
			station := &models.Station{
				ID:            s.ID,
				Name:          s.Name,
				TransportType: s.TransportType,
			}
			cityModel.Stations = append(cityModel.Stations, station)
		}

		if err := h.service.SaveCity(ctx, cityModel); err != nil {
			return fmt.Errorf("failed to save city %s: %w", c.Name, err)
		}
	}

	return nil
}
