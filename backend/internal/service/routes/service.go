package routes

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type Service struct {
	repository GraphRepository
}

type RouteTask struct {
	From *models.Station
	To   *models.Station
}

type GraphRepository interface {
	SaveTrip(ctx context.Context, trip *models.Trip) error
	GetAllStations(ctx context.Context) (map[string]*models.Station, error)
	GetStationDepartures(ctx context.Context, stationID string) ([]models.Trip, error)
	HasConnection(ctx context.Context, fromUID, toUID string) (bool, error)
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

func (s *Service) GetAllStations(ctx context.Context) (map[string]*models.Station, error) {
	return s.repository.GetAllStations(ctx)
}

func (s *Service) GetStationDepartures(ctx context.Context, stationID string) ([]models.Trip, error) {
	return s.repository.GetStationDepartures(ctx, stationID)
}

func (s *Service) CheckConnection(ctx context.Context, from, to string) (bool, error) {
	return s.repository.HasConnection(ctx, from, to)
}

func (s *Service) GetParsingTasks(ctx context.Context) ([]RouteTask, error) {
	stations, err := s.repository.GetAllStations(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.GetAllStations: %w", err)
	}

	var tasks []RouteTask

	for _, fromStation := range stations {
		for _, toID := range fromStation.ConnectedStationsID {
			if toStation, ok := stations[toID]; ok {
				tasks = append(tasks, RouteTask{
					From: fromStation,
					To:   toStation,
				})
			}
		}
	}

	return tasks, nil
}
