package services

import (
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
)

// GuestService handles guest operations
type GuestService struct {
	guestRepo interfaces.GuestRepository
}

// NewGuestService creates a new guest service
func NewGuestService(guestRepo interfaces.GuestRepository) *GuestService {
	return &GuestService{guestRepo: guestRepo}
}

// RegisterGuest creates a new guest
func (s *GuestService) RegisterGuest(id, name, email, phone, idProof string) (*models.Guest, error) {
	guest := models.NewGuest(id, name, email, phone, idProof)
	if err := s.guestRepo.Create(guest); err != nil {
		return nil, err
	}
	return guest, nil
}

