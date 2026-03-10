package services

import (
	"elevator-system/internal/interfaces"
	"elevator-system/internal/models"
	"log"
	"sync"
	"time"
)

// ElevatorService manages a single elevator's operation.
// State Pattern: Handles state transitions (Idle, MovingUp, MovingDown, DoorOpen, etc.)
// SOLID-SRP: Single responsibility - elevator operation logic.
type ElevatorService struct {
	elevator   *models.Elevator
	building   *models.Building
	strategy   interfaces.SchedulingStrategy
	stopCh     chan struct{}
	requestCh  chan *models.Request
	doneCh     chan struct{}
	peakMode   bool
	mu         sync.RWMutex
}

// NewElevatorService creates a new elevator service.
func NewElevatorService(elevator *models.Elevator, building *models.Building, strategy interfaces.SchedulingStrategy) *ElevatorService {
	return &ElevatorService{
		elevator:  elevator,
		building:  building,
		strategy:  strategy,
		stopCh:    make(chan struct{}),
		requestCh: make(chan *models.Request, 100),
		doneCh:    make(chan struct{}),
	}
}

// Start begins the elevator's operation loop (run as goroutine).
func (s *ElevatorService) Start() {
	go s.run()
}

// Stop signals the elevator to stop.
func (s *ElevatorService) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// SubmitRequest adds a request to this elevator's queue.
func (s *ElevatorService) SubmitRequest(req *models.Request) {
	select {
	case s.requestCh <- req:
	default:
		log.Printf("Elevator %s: request channel full, dropping request", s.elevator.ID)
	}
}

// run is the main elevator loop.
func (s *ElevatorService) run() {
	defer close(s.doneCh)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case req := <-s.requestCh:
			s.elevator.AddRequest(req)
		case <-ticker.C:
			s.processNextStep()
		}
	}
}

// processNextStep executes one step of elevator logic.
func (s *ElevatorService) processNextStep() {
	s.elevator.Lock()
	state := s.elevator.State
	s.elevator.Unlock()

	// Edge case: Emergency stop - do nothing
	if state == models.StateEmergencyStop {
		return
	}

	// Edge case: Maintenance
	if state == models.StateMaintenance {
		return
	}

	// Edge case: Door open - wait for door to close
	if state == models.StateDoorOpen {
		// Simulated: door closes after duration (handled in arriveAtFloor)
		return
	}

	queue := s.elevator.GetRequestQueue()
	if len(queue) == 0 {
		s.elevator.SetState(models.StateIdle)
		s.elevator.SetDirection(models.DirectionIdle)
		return
	}

	// Apply scheduling strategy
	ordered := s.strategy.OrderRequests(s.elevator, s.building)
	s.elevator.SetRequestQueue(ordered)

	// Get next floor to visit
	nextFloor := s.getNextFloorToVisit()
	if nextFloor < 0 {
		return
	}

	currentFloor := s.elevator.GetStatus().CurrentFloor
	if nextFloor == currentFloor {
		s.arriveAtFloor(currentFloor)
		return
	}

	// Move toward next floor
	if nextFloor > currentFloor {
		s.elevator.SetState(models.StateMovingUp)
		s.elevator.SetDirection(models.DirectionUp)
		s.elevator.SetFloor(currentFloor + 1)
	} else {
		s.elevator.SetState(models.StateMovingDown)
		s.elevator.SetDirection(models.DirectionDown)
		s.elevator.SetFloor(currentFloor - 1)
	}
}

func (s *ElevatorService) getNextFloorToVisit() int {
	queue := s.elevator.GetRequestQueue()
	if len(queue) == 0 {
		return -1
	}
	current := s.elevator.GetStatus().CurrentFloor
	dir := s.elevator.GetStatus().Direction

	var bestFloor int = -1
	bestDist := 999999

	for _, req := range queue {
		for _, floor := range []int{req.SourceFloor, req.DestFloor} {
			if floor == current {
				continue
			}
			dist := 999999
			if dir == models.DirectionUp {
				if floor > current {
					dist = floor - current
				}
			} else if dir == models.DirectionDown {
				if floor < current {
					dist = current - floor
				}
			} else {
				dist = abs(floor - current)
			}
			if dist < bestDist {
				bestDist = dist
				bestFloor = floor
			}
		}
	}
	return bestFloor
}

func (s *ElevatorService) arriveAtFloor(floor int) {
	s.elevator.SetState(models.StateDoorOpen)
	s.elevator.SetFloor(floor)

	// Simulate door open/close
	time.Sleep(s.elevator.DoorOpenDuration)

	// Serve pickup requests (at source floor) - add load, request stays in queue
	s.elevator.ProcessPickupAtFloor(floor, 70)

	// Serve dropoff requests (at dest floor) - remove load and request
	s.elevator.RemoveDropoffRequestsAtFloor(floor, 70)

	// Close door - transition to moving/idle
	queue := s.elevator.GetRequestQueue()
	if len(queue) == 0 {
		s.elevator.SetState(models.StateIdle)
		s.elevator.SetDirection(models.DirectionIdle)
	} else {
		dir := s.elevator.GetStatus().Direction
		if dir == models.DirectionUp {
			s.elevator.SetState(models.StateMovingUp)
		} else {
			s.elevator.SetState(models.StateMovingDown)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// EmergencyStop stops the elevator immediately.
func (s *ElevatorService) EmergencyStop() {
	s.elevator.SetState(models.StateEmergencyStop)
	s.elevator.SetDirection(models.DirectionIdle)
}

// Resume resumes after emergency/maintenance.
func (s *ElevatorService) Resume() {
	s.elevator.SetState(models.StateIdle)
}
