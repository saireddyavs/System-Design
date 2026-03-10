package interfaces

import "library-management-system/internal/models"

// ReservationRepository defines data access for reservations
type ReservationRepository interface {
	Create(reservation *models.Reservation) error
	GetByID(id string) (*models.Reservation, error)
	Update(reservation *models.Reservation) error
	GetPendingByBookID(bookID string) ([]*models.Reservation, error)
	GetByMemberID(memberID string) ([]*models.Reservation, error)
	ListAll() ([]*models.Reservation, error)
}
