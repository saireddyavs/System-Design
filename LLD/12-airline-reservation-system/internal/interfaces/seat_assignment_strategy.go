package interfaces

import "airline-reservation-system/internal/models"

// SeatAssignmentStrategy defines the contract for seat assignment algorithms (Strategy Pattern)
type SeatAssignmentStrategy interface {
	AssignSeats(seats []*models.Seat, count int, preferredClass models.SeatClass) ([]string, error)
}
