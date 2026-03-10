package interfaces

import "ride-sharing-service/internal/models"

// RiderRepository defines data access operations for riders (Repository pattern)
type RiderRepository interface {
	Create(rider *models.Rider) error
	GetByID(id string) (*models.Rider, error)
	Update(rider *models.Rider) error
}
