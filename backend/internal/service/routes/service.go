package routes

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type Service struct {
	repository GraphRepository
}

type GraphRepository interface {
	SaveTrip(ctx context.Context, trip *models.Trip) error
}

func New(repository GraphRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) SaveTrip(ctx context.Context, trip *models.Trip) error {
	if err := s.repository.SaveTrip(ctx, trip); err != nil {
		return fmt.Errorf("repository.SaveTrip: %w", err)
	}
	return nil
}
