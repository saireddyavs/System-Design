package models

import (
	"fmt"
	"sync"
	"time"
)

var requestIDCounter int
var requestIDMu sync.Mutex

func nextRequestID() string {
	requestIDMu.Lock()
	defer requestIDMu.Unlock()
	requestIDCounter++
	return fmt.Sprintf("REQ-%d-%d", time.Now().UnixNano(), requestIDCounter)
}

// Request represents an elevator request (Command Pattern).
// Encapsulates all information needed to fulfill a request.
// SOLID-SRP: Single responsibility - represents request data.
// Command Pattern: Request acts as a command object.
type Request struct {
	ID          string
	Type        RequestType
	SourceFloor int
	DestFloor   int
	Direction   Direction
	Timestamp   time.Time
	Priority    int // Higher = more urgent (e.g., emergency)
}

// NewExternalRequest creates a request from floor button press.
func NewExternalRequest(sourceFloor int, direction Direction) *Request {
	return &Request{
		ID:          nextRequestID(),
		Type:        RequestTypeExternal,
		SourceFloor: sourceFloor,
		DestFloor:   sourceFloor, // Will be set when passenger selects floor
		Direction:   direction,
		Timestamp:   time.Now(),
		Priority:    0,
	}
}

// NewInternalRequest creates a request from inside elevator (destination selection).
func NewInternalRequest(sourceFloor, destFloor int) *Request {
	dir := DirectionIdle
	if destFloor > sourceFloor {
		dir = DirectionUp
	} else if destFloor < sourceFloor {
		dir = DirectionDown
	}
	return &Request{
		ID:          nextRequestID(),
		Type:        RequestTypeInternal,
		SourceFloor: sourceFloor,
		DestFloor:   destFloor,
		Direction:   dir,
		Timestamp:   time.Now(),
		Priority:    0,
	}
}

// IsPickupRequest returns true if this request is for picking up a passenger.
func (r *Request) IsPickupRequest() bool {
	return r.Type == RequestTypeExternal || r.SourceFloor == r.DestFloor
}
