package cities

import "context"

type Service struct {
	repository GraphRepository
}

func New(repository GraphRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCities(ctx context.Context) ([]string, error) {
	return s.repository.GetAllCities(ctx)
}
