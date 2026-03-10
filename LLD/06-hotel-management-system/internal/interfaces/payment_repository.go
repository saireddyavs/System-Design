package interfaces

import "hotel-management-system/internal/models"

// PaymentRepository defines data access for payments
type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByID(id string) (*models.Payment, error)
	GetByBookingID(bookingID string) ([]*models.Payment, error)
	Update(payment *models.Payment) error
}
