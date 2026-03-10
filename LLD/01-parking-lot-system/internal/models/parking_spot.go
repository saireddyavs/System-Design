package models

import (
	"sync"
	"time"
)

// ParkingSpot represents a single parking space with size and occupancy state.
// SRP: Only responsible for spot state and vehicle compatibility.
type ParkingSpot struct {
	mu         sync.RWMutex
	ID         string
	LevelID    string
	Size       SpotSize
	Vehicle    Vehicle
	OccupiedAt *time.Time
}

// NewParkingSpot creates a new parking spot with default (empty) state.
func NewParkingSpot(id, levelID string, size SpotSize) *ParkingSpot {
	return &ParkingSpot{
		ID:      id,
		LevelID:  levelID,
		Size:    size,
	}
}

// IsAvailable returns true if no vehicle is parked in this spot.
func (s *ParkingSpot) IsAvailable() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Vehicle == nil
}

// CanFit checks if the spot can accommodate the given vehicle.
// A spot can hold vehicles of its size or smaller (Small < Medium < Large).
func (s *ParkingSpot) CanFit(vehicle Vehicle) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Vehicle != nil {
		return false
	}
	return vehicle.GetRequiredSpotSize() <= s.Size
}

// Park assigns the vehicle to this spot. Returns false if spot is occupied
// or vehicle doesn't fit.
func (s *ParkingSpot) Park(vehicle Vehicle) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vehicle != nil || vehicle.GetRequiredSpotSize() > s.Size {
		return false
	}
	now := time.Now()
	s.Vehicle = vehicle
	s.OccupiedAt = &now
	return true
}

// Unpark removes the vehicle from the spot. Returns the vehicle and
// duration parked, or nil if spot was empty.
func (s *ParkingSpot) Unpark() (Vehicle, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vehicle == nil {
		return nil, 0
	}
	v := s.Vehicle
	var duration time.Duration
	if s.OccupiedAt != nil {
		duration = time.Since(*s.OccupiedAt)
	}
	s.Vehicle = nil
	s.OccupiedAt = nil
	return v, duration
}

// GetVehicle returns the currently parked vehicle (for read-only access).
func (s *ParkingSpot) GetVehicle() Vehicle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Vehicle
}

// GetOccupiedAt returns when the spot was occupied.
func (s *ParkingSpot) GetOccupiedAt() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.OccupiedAt
}
