package interfaces

import "elevator-system/internal/models"

// SchedulingStrategy defines the contract for elevator scheduling algorithms.
// Strategy Pattern: Interchangeable algorithms for request ordering.
// SOLID-OCP: Open for extension (new strategies) without modifying dispatcher.
// SOLID-DIP: Dispatcher depends on abstraction, not concrete strategies.
type SchedulingStrategy interface {
	// Name returns the strategy identifier.
	Name() string

	// OrderRequests reorders the elevator's request queue based on algorithm.
	// SCAN/LOOK: Continue in current direction, serve requests along the way.
	// Returns ordered list of requests to serve.
	OrderRequests(elevator *models.Elevator, building *models.Building) []*models.Request
}
