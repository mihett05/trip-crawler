package cities

import "context"

type GraphRepository interface {
	GetAllCities(ctx context.Context) ([]string, error)
}
