package strategies

import (
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"time"
)

// BasePricingStrategy calculates base price from room and nights
type BasePricingStrategy struct{}

func (s *BasePricingStrategy) CalculatePrice(ctx *interfaces.PricingContext) float64 {
	if ctx == nil || ctx.Room == nil || ctx.Nights <= 0 {
		return 0
	}
	return ctx.Room.GetBasePrice() * float64(ctx.Nights)
}

// SeasonalPricingStrategy applies seasonal multipliers (Strategy pattern)
// Peak season (Jun-Aug, Dec): 1.3x, Off-peak: 0.9x
type SeasonalPricingStrategy struct {
	Next interfaces.PricingStrategy
}

func (s *SeasonalPricingStrategy) CalculatePrice(ctx *interfaces.PricingContext) float64 {
	base := s.Next.CalculatePrice(ctx)
	month := ctx.CheckInDate.Month()
	multiplier := 1.0
	switch month {
	case 6, 7, 8, 12: // Peak
		multiplier = 1.3
	case 1, 2, 11: // Off-peak
		multiplier = 0.9
	}
	return base * multiplier
}

// WeekdayWeekendPricingStrategy applies weekend premium (Sat/Sun)
type WeekdayWeekendPricingStrategy struct {
	Next interfaces.PricingStrategy
}

func (s *WeekdayWeekendPricingStrategy) CalculatePrice(ctx *interfaces.PricingContext) float64 {
	base := s.Next.CalculatePrice(ctx)
	weekendNights := 0
	for d := ctx.CheckInDate; d.Before(ctx.CheckOutDate); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			weekendNights++
		}
	}
	if ctx.Nights == 0 {
		return base
	}
	// Apply weekend premium ratio to preserve seasonal from Next
	weekdayNights := ctx.Nights - weekendNights
	weekendMultiplier := 1.2
	ratio := (float64(weekdayNights) + float64(weekendNights)*weekendMultiplier) / float64(ctx.Nights)
	return base * ratio
}

// LoyaltyDiscountStrategy applies loyalty tier discount
type LoyaltyDiscountStrategy struct {
	Next interfaces.PricingStrategy
}

func (s *LoyaltyDiscountStrategy) CalculatePrice(ctx *interfaces.PricingContext) float64 {
	base := s.Next.CalculatePrice(ctx)
	if ctx.Guest == nil {
		return base
	}
	discount := models.LoyaltyTierDiscounts[ctx.Guest.GetLoyaltyTier()]
	return base * (1 - discount)
}

// CompositePricingStrategy chains multiple strategies
type CompositePricingStrategy struct {
	strategies []interfaces.PricingStrategy
}

// NewCompositePricingStrategy creates a chain: Base -> Seasonal -> Weekday -> Loyalty
func NewCompositePricingStrategy() *CompositePricingStrategy {
	base := &BasePricingStrategy{}
	seasonal := &SeasonalPricingStrategy{Next: base}
	weekday := &WeekdayWeekendPricingStrategy{Next: seasonal}
	loyalty := &LoyaltyDiscountStrategy{Next: weekday}
	return &CompositePricingStrategy{
		strategies: []interfaces.PricingStrategy{loyalty},
	}
}

func (s *CompositePricingStrategy) CalculatePrice(ctx *interfaces.PricingContext) float64 {
	if len(s.strategies) == 0 {
		return 0
	}
	return s.strategies[0].CalculatePrice(ctx)
}
