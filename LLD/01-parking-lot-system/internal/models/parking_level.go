package models

import "sync"

// ParkingLevel represents a floor or section of the parking lot with multiple spots.
// SRP: Manages spot collection and provides level-based queries.
type ParkingLevel struct {
	mu     sync.RWMutex
	ID     string
	Name   string
	Spots  []*ParkingSpot
	spotID map[string]*ParkingSpot
}

// NewParkingLevel creates a level with the given spots.
func NewParkingLevel(id, name string, spots []*ParkingSpot) *ParkingLevel {
	spotID := make(map[string]*ParkingSpot)
	for _, s := range spots {
		spotID[s.ID] = s
	}
	return &ParkingLevel{
		ID:     id,
		Name:   name,
		Spots:  spots,
		spotID: spotID,
	}
}

// GetSpot returns a spot by ID.
func (l *ParkingLevel) GetSpot(spotID string) *ParkingSpot {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.spotID[spotID]
}

// GetAvailableSpots returns spots that can fit the given vehicle.
func (l *ParkingLevel) GetAvailableSpots(vehicle Vehicle) []*ParkingSpot {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var available []*ParkingSpot
	for _, s := range l.Spots {
		if s.CanFit(vehicle) {
			available = append(available, s)
		}
	}
	return available
}

// CountAvailableSpots returns the count of spots that can fit the vehicle.
func (l *ParkingLevel) CountAvailableSpots(vehicle Vehicle) int {
	return len(l.GetAvailableSpots(vehicle))
}
