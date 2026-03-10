package interfaces

import (
	"airline-reservation-system/internal/models"
	"time"
)

// FlightRepository defines the contract for flight data access (Repository Pattern)
type FlightRepository interface {
	// Create adds a new flight
	Create(flight *models.Flight) error
	// GetByID retrieves a flight by ID
	GetByID(id string) (*models.Flight, error)
	// GetByFlightNumber retrieves flights by flight number
	GetByFlightNumber(flightNumber string) ([]*models.Flight, error)
	// Update modifies an existing flight
	Update(flight *models.Flight) error
	// Delete removes a flight
	Delete(id string) error
	// SearchByRoute finds flights between origin and destination on a given date
	SearchByRoute(origin, destination string, date time.Time) ([]*models.Flight, error)
	// GetAll returns all flights (for admin purposes)
	GetAll() ([]*models.Flight, error)
}
