package interfaces

import "ride-sharing-service/internal/models"

// MatchingStrategy defines the algorithm for selecting a driver (Strategy pattern)
type MatchingStrategy interface {
	FindDriver(riderLocation models.Location, drivers []*models.Driver) (*models.Driver, error)
}
