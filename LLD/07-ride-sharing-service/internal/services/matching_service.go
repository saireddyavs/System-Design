package services

import (
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
)

const DefaultSearchRadiusKm = 50.0

// MatchingService handles driver-rider matching using configurable strategy
type MatchingService struct {
	driverRepo     interfaces.DriverRepository
	matchingStrategy interfaces.MatchingStrategy
	searchRadiusKm float64
}

// NewMatchingService creates a new matching service
func NewMatchingService(driverRepo interfaces.DriverRepository, strategy interfaces.MatchingStrategy, searchRadiusKm float64) *MatchingService {
	if searchRadiusKm <= 0 {
		searchRadiusKm = DefaultSearchRadiusKm
	}
	return &MatchingService{
		driverRepo:       driverRepo,
		matchingStrategy: strategy,
		searchRadiusKm:   searchRadiusKm,
	}
}

// FindNearestDriver finds the best matching driver for a rider at the given location
func (s *MatchingService) FindNearestDriver(riderLocation models.Location) (*models.Driver, error) {
	drivers, err := s.driverRepo.GetAvailableDriversNear(riderLocation, s.searchRadiusKm)
	if err != nil {
		return nil, err
	}
	return s.matchingStrategy.FindDriver(riderLocation, drivers)
}
