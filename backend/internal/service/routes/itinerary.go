package routes

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type ItineraryBuilder interface {
	Build(ctx context.Context, points []string, startAt int64, minDays, maxDays int) ([]models.RoutePoint, error)
}
