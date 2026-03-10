package interfaces

import "airline-reservation-system/internal/models"

// PricingStrategy defines the contract for price calculation (Strategy Pattern)
type PricingStrategy interface {
	// CalculatePrice computes the total price for given seats
	// basePrice: flight base price
	// seats: seats being booked
	// Returns total price
	CalculatePrice(basePrice float64, seats []*models.Seat) float64
	// Name returns the strategy name for logging/debugging
	Name() string
}
