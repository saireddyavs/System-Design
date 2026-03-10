package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
)

// AutoAssignFirstAvailable assigns the first N available seats (Strategy Pattern)
type AutoAssignFirstAvailable struct{}

// NewAutoAssignFirstAvailable creates a new auto-assign strategy
func NewAutoAssignFirstAvailable() interfaces.SeatAssignmentStrategy {
	return &AutoAssignFirstAvailable{}
}

func (s *AutoAssignFirstAvailable) AssignSeats(seats []*models.Seat, count int, preferredClass models.SeatClass) ([]string, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	var available []*models.Seat
	for _, seat := range seats {
		if !seat.IsAvailable() {
			continue
		}
		if preferredClass != "" && seat.Class != preferredClass {
			continue
		}
		available = append(available, seat)
	}

	if len(available) < count {
		return nil, errors.New("insufficient available seats")
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = available[i].ID
	}
	return result, nil
}

func (s *AutoAssignFirstAvailable) Name() string {
	return "AutoAssignFirstAvailable"
}
