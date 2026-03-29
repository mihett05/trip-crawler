package cities

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type Service struct {
	repository GraphRepository
}

func New(repository GraphRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCities(ctx context.Context) ([]string, error) {
	cities, err := s.repository.GetAllCities(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.GetAllCities: %w", err)
	}

	return cities, nil
}

func (s *Service) SaveCity(ctx context.Context, city *models.City) error {
	if err := s.repository.SaveCity(ctx, city); err != nil {
		return fmt.Errorf("repository.SaveCity: %w", err)
	}
	return nil
}
