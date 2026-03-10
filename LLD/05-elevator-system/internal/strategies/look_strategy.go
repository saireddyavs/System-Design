package strategies

import (
	"elevator-system/internal/models"
)

// LookStrategy implements the LOOK algorithm.
// Unlike SCAN, LOOK reverses when no more requests ahead - doesn't go to building end.
// Strategy Pattern: Encapsulates LOOK algorithm logic.
type LookStrategy struct{}

// NewLookStrategy creates a new LOOK strategy instance.
func NewLookStrategy() *LookStrategy {
	return &LookStrategy{}
}

// Name returns the strategy identifier.
func (s *LookStrategy) Name() string {
	return "LOOK"
}

// OrderRequests orders requests using LOOK algorithm.
// Reverses direction when no more requests in current direction.
func (s *LookStrategy) OrderRequests(elevator *models.Elevator, building *models.Building) []*models.Request {
	elevator.RLock()
	queue := make([]*models.Request, len(elevator.RequestQueue))
	copy(queue, elevator.RequestQueue)
	currentFloor := elevator.CurrentFloor
	direction := elevator.Direction
	elevator.RUnlock()

	if len(queue) == 0 {
		return queue
	}

	floorSeq := s.buildLookFloorSequence(currentFloor, direction, queue)

	return s.requestsByFirstStop(queue, floorSeq)
}

func (s *LookStrategy) buildLookFloorSequence(current int, dir models.Direction, queue []*models.Request) []int {
	floors := make(map[int]bool)
	for _, req := range queue {
		floors[req.SourceFloor] = true
		floors[req.DestFloor] = true
	}

	if dir == models.DirectionIdle {
		dir = s.nearestDirection(current, floors)
	}

	var seq []int
	if dir == models.DirectionUp {
		maxFloor := current
		for f := range floors {
			if f > maxFloor {
				maxFloor = f
			}
		}
		for f := current; f <= maxFloor; f++ {
			if floors[f] {
				seq = append(seq, f)
			}
		}
		minFloor := maxFloor
		for f := range floors {
			if f < minFloor {
				minFloor = f
			}
		}
		for f := maxFloor - 1; f >= minFloor; f-- {
			if floors[f] {
				seq = append(seq, f)
			}
		}
		for f := minFloor - 1; f >= 0; f-- {
			if floors[f] {
				seq = append(seq, f)
			}
		}
	} else {
		minFloor := current
		for f := range floors {
			if f < minFloor {
				minFloor = f
			}
		}
		for f := current; f >= minFloor; f-- {
			if floors[f] {
				seq = append(seq, f)
			}
		}
		maxFloor := minFloor
		for f := range floors {
			if f > maxFloor {
				maxFloor = f
			}
		}
		for f := minFloor + 1; f <= maxFloor; f++ {
			if floors[f] {
				seq = append(seq, f)
			}
		}
	}
	return seq
}

func (s *LookStrategy) nearestDirection(current int, floors map[int]bool) models.Direction {
	minUp, minDown := 999999, 999999
	for f := range floors {
		if f > current && f-current < minUp {
			minUp = f - current
		}
		if f < current && current-f < minDown {
			minDown = current - f
		}
	}
	if minUp <= minDown {
		return models.DirectionUp
	}
	return models.DirectionDown
}

func (s *LookStrategy) requestsByFirstStop(queue []*models.Request, floorSeq []int) []*models.Request {
	floorOrder := make(map[int]int)
	for i, f := range floorSeq {
		floorOrder[f] = i
	}
	firstStopOrder := make(map[string]int)
	for _, req := range queue {
		order := floorOrder[req.SourceFloor]
		if o2 := floorOrder[req.DestFloor]; req.SourceFloor != req.DestFloor && o2 < order {
			order = o2
		}
		firstStopOrder[req.ID] = order
	}
	result := make([]*models.Request, len(queue))
	copy(result, queue)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			oi, oj := firstStopOrder[result[i].ID], firstStopOrder[result[j].ID]
			if oj < oi {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
