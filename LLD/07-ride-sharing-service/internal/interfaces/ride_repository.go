package interfaces

import "ride-sharing-service/internal/models"

// RideRepository defines data access operations for rides (Repository pattern)
type RideRepository interface {
	Create(ride *models.Ride) error
	GetByID(id string) (*models.Ride, error)
	Update(ride *models.Ride) error
	GetActiveRidesByRider(riderID string) ([]*models.Ride, error)
	GetActiveRidesByDriver(driverID string) ([]*models.Ride, error)
	CountActiveRequests() (int, error)
}
