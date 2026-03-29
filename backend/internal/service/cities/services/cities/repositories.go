package cities

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type GraphRepository interface {
	GetAllCities(ctx context.Context) ([]string, error)
	SaveCity(ctx context.Context, city *models.City) error
}
