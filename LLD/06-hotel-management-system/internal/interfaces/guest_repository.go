package interfaces

import "hotel-management-system/internal/models"

// GuestRepository defines data access for guests (Repository pattern)
type GuestRepository interface {
	Create(guest *models.Guest) error
	GetByID(id string) (*models.Guest, error)
	GetByEmail(email string) (*models.Guest, error)
	GetAll() ([]*models.Guest, error)
	Update(guest *models.Guest) error
	Delete(id string) error
}
