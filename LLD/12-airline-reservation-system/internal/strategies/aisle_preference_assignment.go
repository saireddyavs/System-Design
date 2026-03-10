package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
)

// AislePreferenceAssignment prefers aisle seats, falls back to others (Strategy Pattern)
type AislePreferenceAssignment struct{}

// NewAislePreferenceAssignment creates a new aisle preference strategy
func NewAislePreferenceAssignment() interfaces.SeatAssignmentStrategy {
	return &AislePreferenceAssignment{}
}

func (s *AislePreferenceAssignment) AssignSeats(seats []*models.Seat, count int, preferredClass models.SeatClass) ([]string, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	var aisleSeats, otherSeats []*models.Seat
	for _, seat := range seats {
		if !seat.IsAvailable() {
			continue
		}
		if preferredClass != "" && seat.Class != preferredClass {
			continue
		}
		if seat.IsAisle() {
			aisleSeats = append(aisleSeats, seat)
		} else {
			otherSeats = append(otherSeats, seat)
		}
	}

	// Prefer aisle seats first
	allAvailable := append(aisleSeats, otherSeats...)
	if len(allAvailable) < count {
		return nil, errors.New("insufficient available seats")
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = allAvailable[i].ID
	}
	return result, nil
}

func (s *AislePreferenceAssignment) Name() string {
	return "AislePreference"
}
