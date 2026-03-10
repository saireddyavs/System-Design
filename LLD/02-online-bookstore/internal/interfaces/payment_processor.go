package interfaces

import "online-bookstore/internal/models"

// PaymentProcessor defines the contract for payment processing.
// Strategy pattern: OCP - new payment methods added without modifying existing code.
// Each payment method (CreditCard, PayPal) implements this interface.
type PaymentProcessor interface {
	Process(payment *models.Payment) (*models.Payment, error)
	GetMethodName() string
}
