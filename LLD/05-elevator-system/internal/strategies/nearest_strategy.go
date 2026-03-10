package strategies

import (
	"elevator-system/internal/models"
)

// NearestStrategy serves the nearest floor first.
// Strategy Pattern: Simple alternative for low-traffic scenarios.
type NearestStrategy struct{}

// NewNearestStrategy creates a new Nearest strategy instance.
func NewNearestStrategy() *NearestStrategy {
	return &NearestStrategy{}
}

// Name returns the strategy identifier.
func (s *NearestStrategy) Name() string {
	return "NEAREST"
}

// OrderRequests orders requests by proximity to current floor.
func (s *NearestStrategy) OrderRequests(elevator *models.Elevator, building *models.Building) []*models.Request {
	elevator.RLock()
	queue := make([]*models.Request, len(elevator.RequestQueue))
	copy(queue, elevator.RequestQueue)
	currentFloor := elevator.CurrentFloor
	elevator.RUnlock()

	if len(queue) == 0 {
		return queue
	}

	// Score: distance to nearest stop (source or dest)
	score := func(req *models.Request) int {
		ds := abs(currentFloor - req.SourceFloor)
		dd := abs(currentFloor - req.DestFloor)
		if ds < dd {
			return ds
		}
		return dd
	}

	result := make([]*models.Request, len(queue))
	copy(result, queue)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if score(result[j]) < score(result[i]) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
