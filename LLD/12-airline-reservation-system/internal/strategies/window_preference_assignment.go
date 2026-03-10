package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
)

// WindowPreferenceAssignment prefers window seats, falls back to others (Strategy Pattern)
type WindowPreferenceAssignment struct{}

// NewWindowPreferenceAssignment creates a new window preference strategy
func NewWindowPreferenceAssignment() interfaces.SeatAssignmentStrategy {
	return &WindowPreferenceAssignment{}
}

func (s *WindowPreferenceAssignment) AssignSeats(seats []*models.Seat, count int, preferredClass models.SeatClass) ([]string, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	var windowSeats, otherSeats []*models.Seat
	for _, seat := range seats {
		if !seat.IsAvailable() {
			continue
		}
		if preferredClass != "" && seat.Class != preferredClass {
			continue
		}
		if seat.IsWindow() {
			windowSeats = append(windowSeats, seat)
		} else {
			otherSeats = append(otherSeats, seat)
		}
	}

	// Prefer window seats first
	allAvailable := append(windowSeats, otherSeats...)
	if len(allAvailable) < count {
		return nil, errors.New("insufficient available seats")
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = allAvailable[i].ID
	}
	return result, nil
}

func (s *WindowPreferenceAssignment) Name() string {
	return "WindowPreference"
}
