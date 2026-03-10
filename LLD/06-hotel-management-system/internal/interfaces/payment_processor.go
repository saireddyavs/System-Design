package interfaces

import "hotel-management-system/internal/models"

// PaymentProcessor defines payment processing abstraction (Strategy/Adapter)
// I - Interface Segregation: Focused on payment operations only
// D - Dependency Inversion: Services depend on interface, not concrete payment gateway
type PaymentProcessor interface {
	ProcessPayment(payment *models.Payment) error
	ProcessRefund(payment *models.Payment, amount float64) error
}
