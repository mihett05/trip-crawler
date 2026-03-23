package cities

import (
	"context"
	"fmt"
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
