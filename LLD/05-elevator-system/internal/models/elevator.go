package models

import (
	"sync"
	"time"
)

const (
	DefaultCapacity    = 1000 // kg
	DefaultDoorTime    = 2 * time.Second
	DefaultFloorTime   = 1 * time.Second
	OverweightThreshold = 900 // kg - 90% of capacity
)

// Elevator represents a single elevator unit.
// SOLID-SRP: Encapsulates elevator state and behavior.
// State Pattern: State transitions managed via ElevatorState.
type Elevator struct {
	ID           string
	CurrentFloor int
	Direction    Direction
	State        ElevatorState
	Capacity     int // kg
	CurrentLoad  int // kg
	RequestQueue []*Request
	mu           sync.RWMutex

	// Observer pattern: subscribers for floor arrival
	observers []FloorArrivalObserver

	// Config
	DoorOpenDuration time.Duration
	FloorTravelTime  time.Duration
}

// FloorArrivalObserver interface for Observer pattern.
// SOLID-ISP: Interface segregation - only floor arrival events.
type FloorArrivalObserver interface {
	OnFloorArrival(elevatorID string, floor int)
}

// NewElevator creates a new elevator with default config.
func NewElevator(id string) *Elevator {
	return &Elevator{
		ID:               id,
		CurrentFloor:     0,
		Direction:        DirectionIdle,
		State:           StateIdle,
		Capacity:        DefaultCapacity,
		CurrentLoad:     0,
		RequestQueue:    make([]*Request, 0),
		observers:       make([]FloorArrivalObserver, 0),
		DoorOpenDuration: DefaultDoorTime,
		FloorTravelTime:  DefaultFloorTime,
	}
}

// AddObserver adds a floor arrival observer.
func (e *Elevator) AddObserver(o FloorArrivalObserver) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.observers = append(e.observers, o)
}

// notifyObservers notifies all observers of floor arrival.
func (e *Elevator) notifyObservers(floor int) {
	e.mu.RLock()
	observers := make([]FloorArrivalObserver, len(e.observers))
	copy(observers, e.observers)
	e.mu.RUnlock()

	for _, o := range observers {
		o.OnFloorArrival(e.ID, floor)
	}
}

// GetStatus returns a thread-safe snapshot of elevator status.
func (e *Elevator) GetStatus() ElevatorStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return ElevatorStatus{
		ID:           e.ID,
		CurrentFloor: e.CurrentFloor,
		Direction:    e.Direction,
		State:        e.State,
		CurrentLoad:  e.CurrentLoad,
		Capacity:     e.Capacity,
		QueueLength:  len(e.RequestQueue),
		IsOverweight: e.CurrentLoad >= OverweightThreshold,
	}
}

// ElevatorStatus is a read-only snapshot for display.
type ElevatorStatus struct {
	ID           string
	CurrentFloor int
	Direction    Direction
	State        ElevatorState
	CurrentLoad  int
	Capacity     int
	QueueLength  int
	IsOverweight bool
}

// AddRequest adds a request to the queue (thread-safe).
func (e *Elevator) AddRequest(req *Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.RequestQueue = append(e.RequestQueue, req)
}

// SetState updates elevator state.
func (e *Elevator) SetState(s ElevatorState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.State = s
}

// SetDirection updates elevator direction.
func (e *Elevator) SetDirection(d Direction) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Direction = d
}

// SetFloor updates current floor and notifies observers.
func (e *Elevator) SetFloor(floor int) {
	e.mu.Lock()
	e.CurrentFloor = floor
	e.mu.Unlock()
	e.notifyObservers(floor)
}

// AddLoad adds weight (passenger boarding).
func (e *Elevator) AddLoad(weight int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CurrentLoad += weight
}

// RemoveLoad removes weight (passenger alighting).
func (e *Elevator) RemoveLoad(weight int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.CurrentLoad >= weight {
		e.CurrentLoad -= weight
	}
}

// IsOverweight returns true if elevator is at or over weight limit.
func (e *Elevator) IsOverweight() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.CurrentLoad >= OverweightThreshold
}

// CanAcceptPassenger returns false if overweight.
func (e *Elevator) CanAcceptPassenger(weight int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.CurrentLoad+weight < OverweightThreshold
}

// GetRequestQueue returns a copy of the request queue.
func (e *Elevator) GetRequestQueue() []*Request {
	e.mu.RLock()
	defer e.mu.RUnlock()
	queue := make([]*Request, len(e.RequestQueue))
	copy(queue, e.RequestQueue)
	return queue
}

// ProcessPickupAtFloor handles passenger pickup - adds load for external requests only.
// Request stays in queue until dropoff at dest floor.
func (e *Elevator) ProcessPickupAtFloor(floor int, weightPerPassenger int) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	count := 0
	for _, req := range e.RequestQueue {
		if req.SourceFloor == floor && req.Type == RequestTypeExternal &&
			e.CurrentLoad+weightPerPassenger < OverweightThreshold {
			e.CurrentLoad += weightPerPassenger
			count++
		}
	}
	return count
}

// RemoveDropoffRequestsAtFloor removes requests completed at dest floor and removes load.
func (e *Elevator) RemoveDropoffRequestsAtFloor(floor int, weightPerPassenger int) []*Request {
	e.mu.Lock()
	defer e.mu.Unlock()
	var removed []*Request
	var remaining []*Request
	for _, req := range e.RequestQueue {
		if req.DestFloor == floor {
			removed = append(removed, req)
			if e.CurrentLoad >= weightPerPassenger {
				e.CurrentLoad -= weightPerPassenger
			}
		} else {
			remaining = append(remaining, req)
		}
	}
	e.RequestQueue = remaining
	return removed
}

// RemoveRequestsAtFloor removes requests served at the given floor (legacy, use ProcessPickup/RemoveDropoff).
func (e *Elevator) RemoveRequestsAtFloor(floor int, isPickup bool) []*Request {
	if isPickup {
		e.ProcessPickupAtFloor(floor, 70)
		return nil // Don't remove at pickup
	}
	return e.RemoveDropoffRequestsAtFloor(floor, 70)
}

// SetRequestQueue replaces the queue (used by scheduling strategy).
func (e *Elevator) SetRequestQueue(queue []*Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.RequestQueue = queue
}

// Lock/Unlock for external synchronization.
func (e *Elevator) Lock()    { e.mu.Lock() }
func (e *Elevator) Unlock()  { e.mu.Unlock() }
func (e *Elevator) RLock()   { e.mu.RLock() }
func (e *Elevator) RUnlock() { e.mu.RUnlock() }
