package strategies

import "parking-lot-system/internal/models"

// NearestSpotStrategy selects the first available spot (nearest to entrance).
// Simple strategy: iterates spots in order, returning the first that fits.
// In a real system, "nearest" could factor in level, row, distance from elevators.
type NearestSpotStrategy struct{}

// NewNearestSpotStrategy creates a new nearest-spot strategy.
func NewNearestSpotStrategy() *NearestSpotStrategy {
	return &NearestSpotStrategy{}
}

// FindSpot returns the first spot that can fit the vehicle.
func (s *NearestSpotStrategy) FindSpot(vehicle models.Vehicle, spots []*models.ParkingSpot) *models.ParkingSpot {
	for _, spot := range spots {
		if spot.CanFit(vehicle) {
			return spot
		}
	}
	return nil
}
