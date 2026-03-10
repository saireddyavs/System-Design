package strategies

import (
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"time"
)

// BaseFareStrategy implements base + distance + time fare calculation
type BaseFareStrategy struct {
	BaseFare       float64 // Fixed base fare
	PerKmRate      float64 // Rate per kilometer
	PerMinuteRate  float64 // Rate per minute
}

// NewBaseFareStrategy creates standard fare calculator
func NewBaseFareStrategy(baseFare, perKm, perMin float64) *BaseFareStrategy {
	return &BaseFareStrategy{
		BaseFare:      baseFare,
		PerKmRate:     perKm,
		PerMinuteRate: perMin,
	}
}

// Calculate implements FareCalculator
func (s *BaseFareStrategy) Calculate(pickup, dropoff models.Location, duration time.Duration, surgeMultiplier float64) float64 {
	distance := models.HaversineDistance(pickup, dropoff)
	minutes := duration.Minutes()

	fare := s.BaseFare + (distance * s.PerKmRate) + (minutes * s.PerMinuteRate)
	return fare * surgeMultiplier
}

var _ interfaces.FareCalculator = (*BaseFareStrategy)(nil)

// SurgePricingStrategy wraps another calculator and applies surge multiplier
type SurgePricingStrategy struct {
	BaseCalculator   interfaces.FareCalculator
	SurgeThreshold   int     // Active requests threshold for surge
	SurgeMultiplier  float64 // e.g., 1.5 for 50% surge
}

// NewSurgePricingStrategy creates a fare calculator with surge pricing
func NewSurgePricingStrategy(base interfaces.FareCalculator, threshold int, multiplier float64) *SurgePricingStrategy {
	return &SurgePricingStrategy{
		BaseCalculator:  base,
		SurgeThreshold:  threshold,
		SurgeMultiplier: multiplier,
	}
}

// Calculate applies surge multiplier when active requests exceed threshold
func (s *SurgePricingStrategy) Calculate(pickup, dropoff models.Location, duration time.Duration, surgeMultiplier float64) float64 {
	// Use provided surge multiplier (calculated by service based on demand)
	effectiveMultiplier := surgeMultiplier
	if effectiveMultiplier < 1.0 {
		effectiveMultiplier = 1.0
	}
	return s.BaseCalculator.Calculate(pickup, dropoff, duration, effectiveMultiplier)
}
