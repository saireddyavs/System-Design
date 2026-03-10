package interfaces

import "shopping-cart-system/internal/models"

// PaymentRequest contains payment details
type PaymentRequest struct {
	Amount        float64
	Currency      string
	PaymentMethod models.PaymentMethod
	UserID        string
	OrderID       string
	Metadata      map[string]string
}

// PaymentResult contains payment outcome
type PaymentResult struct {
	Success   bool
	PaymentID string
	Message   string
}

// PaymentProcessor defines payment processing (Strategy pattern)
type PaymentProcessor interface {
	Process(req *PaymentRequest) (*PaymentResult, error)
	Supports(method models.PaymentMethod) bool
}
