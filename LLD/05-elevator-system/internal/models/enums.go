package models

// Direction represents elevator movement direction.
// SOLID-SRP: Single responsibility - defines direction semantics only.
type Direction int

const (
	DirectionIdle Direction = iota
	DirectionUp
	DirectionDown
)

func (d Direction) String() string {
	switch d {
	case DirectionUp:
		return "UP"
	case DirectionDown:
		return "DOWN"
	default:
		return "IDLE"
	}
}

// ElevatorState represents the current state of an elevator.
// State Pattern: Each state encapsulates behavior for that phase.
type ElevatorState int

const (
	StateIdle ElevatorState = iota
	StateMovingUp
	StateMovingDown
	StateDoorOpen
	StateMaintenance
	StateEmergencyStop
)

func (s ElevatorState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateMovingUp:
		return "MOVING_UP"
	case StateMovingDown:
		return "MOVING_DOWN"
	case StateDoorOpen:
		return "DOOR_OPEN"
	case StateMaintenance:
		return "MAINTENANCE"
	case StateEmergencyStop:
		return "EMERGENCY_STOP"
	default:
		return "UNKNOWN"
	}
}

// RequestType distinguishes external (floor button) from internal (destination) requests.
type RequestType int

const (
	RequestTypeExternal RequestType = iota
	RequestTypeInternal
)

func (r RequestType) String() string {
	switch r {
	case RequestTypeExternal:
		return "EXTERNAL"
	case RequestTypeInternal:
		return "INTERNAL"
	default:
		return "UNKNOWN"
	}
}
