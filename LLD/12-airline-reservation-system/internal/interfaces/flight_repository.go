package interfaces

import (
	"airline-reservation-system/internal/models"
	"time"
)

// FlightRepository defines the contract for flight data access (Repository Pattern)
type FlightRepository interface {
	Create(flight *models.Flight) error
	GetByID(id string) (*models.Flight, error)
	Update(flight *models.Flight) error
	SearchByRoute(origin, destination string, date time.Time) ([]*models.Flight, error)
}
