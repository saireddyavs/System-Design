package services

import (
	"errors"
	"elevator-system/internal/interfaces"
	"elevator-system/internal/models"
	"log"
	"sync"
)

// DispatcherService assigns incoming requests to elevators.
// Singleton pattern: Single point of control for request distribution.
// SOLID-SRP: Single responsibility - request routing.
type DispatcherService struct {
	building     *models.Building
	elevatorSvcs map[string]*ElevatorService
	strategy     interfaces.SchedulingStrategy
	requestCh    chan *dispatchRequest
	stopCh       chan struct{}
	doneCh       chan struct{}
	peakMode     bool
	mu           sync.RWMutex
}

type dispatchRequest struct {
	req      *models.Request
	response chan error
}

var (
	dispatcherInstance *DispatcherService
	dispatcherOnce     sync.Once
)

// GetDispatcher returns the singleton dispatcher instance (optional use).
// Singleton Pattern: Ensures single building controller per process.
func GetDispatcher(building *models.Building, strategy interfaces.SchedulingStrategy) *DispatcherService {
	dispatcherOnce.Do(func() {
		dispatcherInstance = NewDispatcherService(building, strategy)
	})
	return dispatcherInstance
}

// NewDispatcherService creates a new dispatcher (for testing).
func NewDispatcherService(building *models.Building, strategy interfaces.SchedulingStrategy) *DispatcherService {
	return &DispatcherService{
		building:     building,
		elevatorSvcs: make(map[string]*ElevatorService),
		strategy:     strategy,
		requestCh:    make(chan *dispatchRequest, 200),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
}

// RegisterElevatorService registers an elevator service with the dispatcher.
func (d *DispatcherService) RegisterElevatorService(svc *ElevatorService) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.elevatorSvcs[svc.elevator.ID] = svc
}

// Start begins the dispatcher's request processing loop.
func (d *DispatcherService) Start() {
	go d.run()
}

// Stop signals the dispatcher to stop.
func (d *DispatcherService) Stop() {
	close(d.stopCh)
	<-d.doneCh
}

// SubmitRequest assigns a request to the best elevator.
func (d *DispatcherService) SubmitRequest(req *models.Request) error {
	respCh := make(chan error, 1)
	select {
	case d.requestCh <- &dispatchRequest{req: req, response: respCh}:
		return <-respCh
	default:
		return ErrRequestQueueFull
	}
}

var ErrRequestQueueFull = errors.New("request queue full")

// run processes incoming requests.
func (d *DispatcherService) run() {
	defer close(d.doneCh)
	for {
		select {
		case <-d.stopCh:
			return
		case dr := <-d.requestCh:
			err := d.assignRequest(dr.req)
			select {
			case dr.response <- err:
			default:
			}
		}
	}
}

// assignRequest selects the best elevator and assigns the request.
func (d *DispatcherService) assignRequest(req *models.Request) error {
	elevator := d.selectElevator(req)
	if elevator == nil {
		log.Printf("No suitable elevator for request from floor %d", req.SourceFloor)
		return errors.New("no suitable elevator")
	}
	elevator.SubmitRequest(req)
	return nil
}

// selectElevator picks the best elevator using SCAN/LOOK-style assignment.
// Chooses nearest elevator moving toward request, or idle elevator.
func (d *DispatcherService) selectElevator(req *models.Request) *ElevatorService {
	d.mu.RLock()
	services := make([]*ElevatorService, 0, len(d.elevatorSvcs))
	for _, svc := range d.elevatorSvcs {
		services = append(services, svc)
	}
	d.mu.RUnlock()

	sourceFloor := req.SourceFloor
	dir := req.Direction

	var best *ElevatorService
	bestScore := 999999

	for _, svc := range services {
		status := svc.elevator.GetStatus()
		if status.State == models.StateEmergencyStop || status.State == models.StateMaintenance {
			continue
		}
		if status.IsOverweight {
			continue
		}

		score := d.computeScore(svc, sourceFloor, dir)
		if score < bestScore {
			bestScore = score
			best = svc
		}
	}
	return best
}

func (d *DispatcherService) computeScore(svc *ElevatorService, floor int, dir models.Direction) int {
	status := svc.elevator.GetStatus()
	dist := abs(status.CurrentFloor - floor)

	// Prefer elevator moving toward request
	if status.Direction == dir {
		if (dir == models.DirectionUp && status.CurrentFloor <= floor) ||
			(dir == models.DirectionDown && status.CurrentFloor >= floor) {
			return dist // Same direction, will pass this floor
		}
	}
	if status.State == models.StateIdle {
		return dist
	}
	// Elevator going wrong way - higher penalty
	return dist + 100
}

// GetStatus returns status of all elevators.
func (d *DispatcherService) GetStatus() []models.ElevatorStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []models.ElevatorStatus
	for _, svc := range d.elevatorSvcs {
		result = append(result, svc.elevator.GetStatus())
	}
	return result
}

// EmergencyStop stops a specific elevator.
func (d *DispatcherService) EmergencyStop(elevatorID string) error {
	d.mu.RLock()
	svc, ok := d.elevatorSvcs[elevatorID]
	d.mu.RUnlock()
	if !ok {
		return errors.New("elevator not found")
	}
	svc.EmergencyStop()
	return nil
}

// ResumeElevator resumes an elevator.
func (d *DispatcherService) ResumeElevator(elevatorID string) error {
	d.mu.RLock()
	svc, ok := d.elevatorSvcs[elevatorID]
	d.mu.RUnlock()
	if !ok {
		return errors.New("elevator not found")
	}
	svc.Resume()
	return nil
}

// SetPeakMode enables peak period handling.
func (d *DispatcherService) SetPeakMode(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.peakMode = enabled
}