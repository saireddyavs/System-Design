package interfaces

import "ride-sharing-service/internal/models"

// DriverRepository defines data access operations for drivers (Repository pattern)
type DriverRepository interface {
	Create(driver *models.Driver) error
	GetByID(id string) (*models.Driver, error)
	Update(driver *models.Driver) error
	GetAvailableDrivers() ([]*models.Driver, error)
	GetAvailableDriversNear(location models.Location, radiusKm float64) ([]*models.Driver, error)
}
