package interfaces

import (
	"parking-lot-system/internal/models"
)

// ParkingStrategy defines how to select a spot from available candidates.
// Strategy pattern + ISP: Small focused interface - implementers only need
// to provide spot selection logic. OCP: New strategies (e.g., cheapest-first,
// disabled-preferential) can be added without changing ParkingService.
type ParkingStrategy interface {
	// FindSpot returns the best spot from available spots for the vehicle.
	// Returns nil if no suitable spot exists.
	FindSpot(vehicle models.Vehicle, spots []*models.ParkingSpot) *models.ParkingSpot
}
