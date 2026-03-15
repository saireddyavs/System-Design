package interfaces

import "movie-ticket-booking/internal/models"

// BookingRepository defines operations for booking data access (Repository Pattern)
type BookingRepository interface {
	Create(booking *models.Booking) error
	GetByID(id string) (*models.Booking, error)
	Update(booking *models.Booking) error
}

// UserRepository defines operations for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
}
