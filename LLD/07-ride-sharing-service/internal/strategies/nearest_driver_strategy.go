package strategies

import (
	"errors"
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"sort"
)

const (
	// DefaultMaxSearchRadiusKm is the maximum radius to search for drivers
	DefaultMaxSearchRadiusKm = 50.0
)

var ErrNoDriverAvailable = errors.New("no driver available within search radius")

// NearestDriverStrategy finds the nearest available driver (Strategy pattern)
type NearestDriverStrategy struct {
	MaxSearchRadiusKm float64
}

// NewNearestDriverStrategy creates a new nearest driver matching strategy
func NewNearestDriverStrategy(maxRadiusKm float64) *NearestDriverStrategy {
	if maxRadiusKm <= 0 {
		maxRadiusKm = DefaultMaxSearchRadiusKm
	}
	return &NearestDriverStrategy{MaxSearchRadiusKm: maxRadiusKm}
}

// FindDriver implements MatchingStrategy - selects nearest driver within radius
func (s *NearestDriverStrategy) FindDriver(riderLocation models.Location, drivers []*models.Driver) (*models.Driver, error) {
	if len(drivers) == 0 {
		return nil, ErrNoDriverAvailable
	}

	// Filter available drivers within radius and with minimum rating
	type driverDistance struct {
		driver   *models.Driver
		distance float64
	}

	var candidates []driverDistance
	for _, d := range drivers {
		if !d.IsAvailable() {
			continue
		}
		// Business rule: Driver rating below 3.0 is deactivated (handled at registration/update)
		if d.GetRating() < 3.0 {
			continue
		}
		dist := models.HaversineDistance(riderLocation, d.GetLocation())
		if dist <= s.MaxSearchRadiusKm {
			candidates = append(candidates, driverDistance{driver: d, distance: dist})
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoDriverAvailable
	}

	// Sort by distance (nearest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	return candidates[0].driver, nil
}

// Ensure NearestDriverStrategy implements MatchingStrategy
var _ interfaces.MatchingStrategy = (*NearestDriverStrategy)(nil)
