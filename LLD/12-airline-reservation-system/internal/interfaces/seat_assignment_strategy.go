package interfaces

import "airline-reservation-system/internal/models"

// SeatAssignmentStrategy defines the contract for seat assignment algorithms (Strategy Pattern)
type SeatAssignmentStrategy interface {
	// AssignSeats selects seats from available seats based on the strategy
	// count: number of seats to assign
	// preferredClass: optional class preference, empty means any
	// Returns selected seat IDs or error if insufficient seats
	AssignSeats(seats []*models.Seat, count int, preferredClass models.SeatClass) ([]string, error)
	// Name returns the strategy name for logging/debugging
	Name() string
}
