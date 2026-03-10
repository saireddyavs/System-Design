package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
)

// DemandBasedPricing adds demand multiplier based on occupancy (Strategy Pattern)
// Higher occupancy = higher price (e.g., 80% full = 1.2x, 90% full = 1.4x)
type DemandBasedPricing struct {
	BaseStrategy interfaces.PricingStrategy
}

// NewDemandBasedPricing wraps a base pricing strategy with demand-based adjustment
func NewDemandBasedPricing(base interfaces.PricingStrategy) interfaces.PricingStrategy {
	return &DemandBasedPricing{BaseStrategy: base}
}

func (p *DemandBasedPricing) CalculatePrice(basePrice float64, seats []*models.Seat) float64 {
	baseTotal := p.BaseStrategy.CalculatePrice(basePrice, seats)

	// Calculate occupancy: count booked seats
	totalSeats := len(seats)
	if totalSeats == 0 {
		return 0
	}
	bookedCount := 0
	for _, s := range seats {
		if s.Status == models.SeatStatusBooked {
			bookedCount++
		}
	}
	occupancy := float64(bookedCount) / float64(totalSeats)

	// Demand multiplier: 0-50% = 1.0, 50-80% = 1.1, 80-90% = 1.2, 90%+ = 1.4
	demandMultiplier := 1.0
	switch {
	case occupancy >= 0.9:
		demandMultiplier = 1.4
	case occupancy >= 0.8:
		demandMultiplier = 1.2
	case occupancy >= 0.5:
		demandMultiplier = 1.1
	}

	return baseTotal * demandMultiplier
}

func (p *DemandBasedPricing) Name() string {
	return "DemandBased"
}
