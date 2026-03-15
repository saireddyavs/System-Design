package interfaces

import "airline-reservation-system/internal/models"

// PricingStrategy defines the contract for price calculation (Strategy Pattern)
type PricingStrategy interface {
	CalculatePrice(basePrice float64, seats []*models.Seat) float64
}
