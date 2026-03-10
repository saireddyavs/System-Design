package interfaces

import (
	"hotel-management-system/internal/models"
	"time"
)

// BookingRepository defines data access for bookings (Repository pattern)
type BookingRepository interface {
	Create(booking *models.Booking) error
	GetByID(id string) (*models.Booking, error)
	GetByGuestID(guestID string) ([]*models.Booking, error)
	GetByRoomID(roomID string) ([]*models.Booking, error)
	Update(booking *models.Booking) error
	Delete(id string) error

	// Availability and overlap checks
	GetBookingsForRoomInRange(roomID string, checkIn, checkOut time.Time) ([]*models.Booking, error)
	GetActiveBookingsForRoom(roomID string, asOf time.Time) ([]*models.Booking, error)
}
