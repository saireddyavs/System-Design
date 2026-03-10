package interfaces

import "airline-reservation-system/internal/models"

// BookingRepository defines the contract for booking data access (Repository Pattern)
type BookingRepository interface {
	// Create adds a new booking
	Create(booking *models.Booking) error
	// GetByID retrieves a booking by ID
	GetByID(id string) (*models.Booking, error)
	// GetByBookingRef retrieves a booking by booking reference
	GetByBookingRef(bookingRef string) (*models.Booking, error)
	// GetByPassengerID retrieves all bookings for a passenger
	GetByPassengerID(passengerID string) ([]*models.Booking, error)
	// GetByFlightID retrieves all bookings for a flight
	GetByFlightID(flightID string) ([]*models.Booking, error)
	// Update modifies an existing booking
	Update(booking *models.Booking) error
	// Delete removes a booking
	Delete(id string) error
}
