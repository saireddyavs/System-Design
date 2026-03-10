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

// GetGuest returns guest by ID
func (s *GuestService) GetGuest(id string) (*models.Guest, error) {
	guest, err := s.guestRepo.GetByID(id)
	if err != nil {
		return nil, ErrGuestNotFound
	}
	return guest, nil
}

// AddLoyaltyPoints adds points to guest
func (s *GuestService) AddLoyaltyPoints(guestID string, points int) error {
	guest, err := s.guestRepo.GetByID(guestID)
	if err != nil {
		return ErrGuestNotFound
	}
	guest.AddLoyaltyPoints(points)
	return s.guestRepo.Update(guest)
}
