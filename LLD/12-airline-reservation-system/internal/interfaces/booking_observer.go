package interfaces

import "airline-reservation-system/internal/models"

// BookingObserver defines the contract for booking event notifications (Observer Pattern)
type BookingObserver interface {
	// OnBookingCreated is called when a new booking is created
	OnBookingCreated(booking *models.Booking)
	// OnBookingCancelled is called when a booking is cancelled
	OnBookingCancelled(booking *models.Booking)
}
