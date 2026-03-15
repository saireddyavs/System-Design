package services

import (
	"errors"
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
)

var ErrRiderAlreadyExists = errors.New("rider already exists")

// RiderService handles rider business logic
type RiderService struct {
	riderRepo interfaces.RiderRepository
}

// NewRiderService creates a new rider service
func NewRiderService(riderRepo interfaces.RiderRepository) *RiderService {
	return &RiderService{riderRepo: riderRepo}
}

// RegisterRider registers a new rider
func (s *RiderService) RegisterRider(id, name, phone, email string) (*models.Rider, error) {
	existing, _ := s.riderRepo.GetByID(id)
	if existing != nil {
		return nil, ErrRiderAlreadyExists
	}
	rider := models.NewRider(id, name, phone, email)
	if err := s.riderRepo.Create(rider); err != nil {
		return nil, err
	}
	return rider, nil
}

// UpdateLocation updates rider's location
func (s *RiderService) UpdateLocation(riderID string, location models.Location) error {
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return err
	}
	rider.UpdateLocation(location)
	return s.riderRepo.Update(rider)
}

