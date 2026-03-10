package strategies

import (
	"elevator-system/internal/models"
)

// ScanStrategy implements the SCAN (elevator) algorithm.
// Elevator moves in one direction to the end of the building, then reverses.
// Serves all requests along the way. Even when no requests, goes to end.
// Strategy Pattern: Encapsulates SCAN algorithm logic.
type ScanStrategy struct{}

// NewScanStrategy creates a new SCAN strategy instance.
func NewScanStrategy() *ScanStrategy {
	return &ScanStrategy{}
}

// Name returns the strategy identifier.
func (s *ScanStrategy) Name() string {
	return "SCAN"
}

// OrderRequests orders requests using SCAN algorithm.
// 1. Continue in current direction to end of building
// 2. Serve all requests in path
// 3. Reverse at end and repeat
func (s *ScanStrategy) OrderRequests(elevator *models.Elevator, building *models.Building) []*models.Request {
	elevator.RLock()
	queue := make([]*models.Request, len(elevator.RequestQueue))
	copy(queue, elevator.RequestQueue)
	currentFloor := elevator.CurrentFloor
	direction := elevator.Direction
	elevator.RUnlock()

	if len(queue) == 0 {
		return queue
	}

	// Build ordered floor sequence using SCAN
	floorSeq := s.buildScanFloorSequence(currentFloor, direction, building.TotalFloors, queue)

	// Map requests to their first stop in sequence order
	return s.requestsByFirstStop(queue, floorSeq)
}

func (s *ScanStrategy) buildScanFloorSequence(current int, dir models.Direction, maxFloor int, queue []*models.Request) []int {
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
		for f := current; f <= maxFloor; f++ {
			if floors[f] {
				seq = append(seq, f)
			}
		}
		for f := maxFloor - 1; f >= 0; f-- {
			if floors[f] {
				seq = append(seq, f)
			}
		}
	} else {
		for f := current; f >= 0; f-- {
			if floors[f] {
				seq = append(seq, f)
			}
		}
		for f := 1; f <= maxFloor; f++ {
			if floors[f] {
				seq = append(seq, f)
			}
		}
	}
	return seq
}

func (s *ScanStrategy) nearestDirection(current int, floors map[int]bool) models.Direction {
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

func (s *ScanStrategy) requestsByFirstStop(queue []*models.Request, floorSeq []int) []*models.Request {
	floorOrder := make(map[int]int)
	for i, f := range floorSeq {
		floorOrder[f] = i
	}
	// Sort requests by first stop (min of source and dest) in floor sequence order
	firstStopOrder := make(map[string]int)
	for _, req := range queue {
		order := floorOrder[req.SourceFloor]
		if o2 := floorOrder[req.DestFloor]; req.SourceFloor != req.DestFloor && o2 < order {
			order = o2
		}
		firstStopOrder[req.ID] = order
	}
	// Sort queue by firstStopOrder
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
