package interfaces

import "airline-reservation-system/internal/models"

// PassengerRepository defines the contract for passenger data access
type PassengerRepository interface {
	Create(passenger *models.Passenger) error
	GetByID(id string) (*models.Passenger, error)
	Update(passenger *models.Passenger) error
	Delete(id string) error
}
