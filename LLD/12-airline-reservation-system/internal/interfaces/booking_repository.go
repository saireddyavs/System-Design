package interfaces

import "airline-reservation-system/internal/models"

// BookingRepository defines the contract for booking data access (Repository Pattern)
type BookingRepository interface {
	Create(booking *models.Booking) error
	GetByID(id string) (*models.Booking, error)
	GetByBookingRef(bookingRef string) (*models.Booking, error)
	GetByFlightID(flightID string) ([]*models.Booking, error)
	Update(booking *models.Booking) error
}
