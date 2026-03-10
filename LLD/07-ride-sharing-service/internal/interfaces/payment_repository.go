package interfaces

import "ride-sharing-service/internal/models"

// PaymentRepository defines data access operations for payments
type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByID(id string) (*models.Payment, error)
	GetByRideID(rideID string) (*models.Payment, error)
	Update(payment *models.Payment) error
}
