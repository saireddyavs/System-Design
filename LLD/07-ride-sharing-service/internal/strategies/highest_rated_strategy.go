package strategies

import (
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"sort"
)

// HighestRatedStrategy finds the highest-rated available driver within radius
type HighestRatedStrategy struct {
	MaxSearchRadiusKm float64
}

// NewHighestRatedStrategy creates a new highest-rated driver matching strategy
func NewHighestRatedStrategy(maxRadiusKm float64) *HighestRatedStrategy {
	if maxRadiusKm <= 0 {
		maxRadiusKm = DefaultMaxSearchRadiusKm
	}
	return &HighestRatedStrategy{MaxSearchRadiusKm: maxRadiusKm}
}

// FindDriver implements MatchingStrategy - selects highest-rated driver within radius
func (s *HighestRatedStrategy) FindDriver(riderLocation models.Location, drivers []*models.Driver) (*models.Driver, error) {
	if len(drivers) == 0 {
		return nil, ErrNoDriverAvailable
	}

	type driverScore struct {
		driver   *models.Driver
		distance float64
	}

	var candidates []driverScore
	for _, d := range drivers {
		if !d.IsAvailable() || d.GetRating() < 3.0 {
			continue
		}
		dist := models.HaversineDistance(riderLocation, d.GetLocation())
		if dist <= s.MaxSearchRadiusKm {
			candidates = append(candidates, driverScore{driver: d, distance: dist})
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoDriverAvailable
	}

	// Sort by rating (desc), then by distance (asc) as tiebreaker
	sort.Slice(candidates, func(i, j int) bool {
		ri, rj := candidates[i].driver.GetRating(), candidates[j].driver.GetRating()
		if ri != rj {
			return ri > rj
		}
		return candidates[i].distance < candidates[j].distance
	})

	return candidates[0].driver, nil
}

var _ interfaces.MatchingStrategy = (*HighestRatedStrategy)(nil)
