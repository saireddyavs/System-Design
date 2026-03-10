package interfaces

import "shopping-cart-system/internal/models"

// CheckoutSummary represents the built checkout result (Builder pattern)
type CheckoutSummary struct {
	OrderID     string
	Subtotal    float64
	Discount    float64
	Tax         float64
	Total       float64
	Order       *models.Order
	PaymentID   string
}
