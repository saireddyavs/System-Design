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
