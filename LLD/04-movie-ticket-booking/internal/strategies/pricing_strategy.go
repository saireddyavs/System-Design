package strategies

import (
	"movie-ticket-booking/internal/models"
	"time"
)

// PricingStrategy defines price calculation (Strategy Pattern - OCP)
type PricingStrategy interface {
	CalculatePrice(basePrice float64, seatCategory models.SeatCategory, showTime time.Time) float64
}

// WeekdayPricingStrategy - lower prices on weekdays
type WeekdayPricingStrategy struct {
	WeekendMultiplier float64
	CategoryMultiplier map[models.SeatCategory]float64
}

// NewWeekdayPricingStrategy creates default pricing strategy
func NewWeekdayPricingStrategy() *WeekdayPricingStrategy {
	return &WeekdayPricingStrategy{
		WeekendMultiplier: 1.25,
		CategoryMultiplier: map[models.SeatCategory]float64{
			models.SeatCategoryRegular: 1.0,
			models.SeatCategoryPremium: 1.5,
			models.SeatCategoryVIP:     2.0,
		},
	}
}

// CalculatePrice computes final price based on day and seat category
func (s *WeekdayPricingStrategy) CalculatePrice(basePrice float64, seatCategory models.SeatCategory, showTime time.Time) float64 {
	price := basePrice

	// Weekend multiplier (Saturday=6, Sunday=0)
	weekday := showTime.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		price *= s.WeekendMultiplier
	}

	// Seat category multiplier
	if mult, ok := s.CategoryMultiplier[seatCategory]; ok {
		price *= mult
	}

	return price
}

// FlatPricingStrategy - simple flat pricing
type FlatPricingStrategy struct {
	CategoryMultiplier map[models.SeatCategory]float64
}

// NewFlatPricingStrategy creates flat pricing
func NewFlatPricingStrategy() *FlatPricingStrategy {
	return &FlatPricingStrategy{
		CategoryMultiplier: map[models.SeatCategory]float64{
			models.SeatCategoryRegular: 1.0,
			models.SeatCategoryPremium: 1.5,
			models.SeatCategoryVIP:     2.0,
		},
	}
}

// CalculatePrice returns price with only category multiplier
func (s *FlatPricingStrategy) CalculatePrice(basePrice float64, seatCategory models.SeatCategory, _ time.Time) float64 {
	mult := 1.0
	if m, ok := s.CategoryMultiplier[seatCategory]; ok {
		mult = m
	}
	return basePrice * mult
}
