package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
)

// ClassMultiplierPricing applies base price + class multiplier (Strategy Pattern)
// Economy 1x, Business 2.5x, First 5x
type ClassMultiplierPricing struct{}

// NewClassMultiplierPricing creates a new class multiplier pricing strategy
func NewClassMultiplierPricing() interfaces.PricingStrategy {
	return &ClassMultiplierPricing{}
}

func (p *ClassMultiplierPricing) CalculatePrice(basePrice float64, seats []*models.Seat) float64 {
	var total float64
	for _, seat := range seats {
		price := basePrice * seat.Class.ClassMultiplier()
		total += price
	}
	return total
}
