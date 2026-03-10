package interfaces

import (
	"ride-sharing-service/internal/models"
)

// PaymentProcessor defines payment processing operations
type PaymentProcessor interface {
	ProcessPayment(rideID string, amount float64, method models.PaymentMethod) (*models.Payment, error)
	RefundPayment(paymentID string) error
}
