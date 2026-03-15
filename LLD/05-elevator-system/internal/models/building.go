package models

import (
	"fmt"
	"sync"
)

// Building represents the physical building with floors and elevators.
// SOLID-SRP: Encapsulates building structure and elevator collection.
type Building struct {
	ID          string
	Name        string
	TotalFloors int
	Elevators   []*Elevator
	mu          sync.RWMutex
}

// NewBuilding creates a new building with specified elevators.
func NewBuilding(id, name string, totalFloors int, elevatorCount int) *Building {
	elevators := make([]*Elevator, elevatorCount)
	for i := 0; i < elevatorCount; i++ {
		elevators[i] = NewElevator(fmt.Sprintf("%s-E%d", id, i))
	}
	return &Building{
		ID:          id,
		Name:        name,
		TotalFloors: totalFloors,
		Elevators:   elevators,
	}
}

// GetElevators returns a copy of elevator slice.
func (b *Building) GetElevators() []*Elevator {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]*Elevator{}, b.Elevators...)
}
