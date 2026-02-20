package service

import (
	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/repository"
)

type HealthService interface {
	Health() error
}

type healthService struct {
	repo repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) HealthService {
	return &healthService{repo: repo}
}

func (s *healthService) Health() error {
	if err := s.repo.Ping(); err != nil {
		return domain.NewInternal("database health check failed: " + err.Error())
	}
	return nil
}
