package models

import (
	"sync"
	"time"
)

// RideStatus represents the state of a ride (State pattern)
type RideStatus string

const (
	RideStatusRequested    RideStatus = "requested"
	RideStatusDriverAssigned RideStatus = "driver_assigned"
	RideStatusInProgress   RideStatus = "in_progress"
	RideStatusCompleted   RideStatus = "completed"
	RideStatusCancelled    RideStatus = "cancelled"
)

// Ride represents a ride-sharing trip
type Ride struct {
	ID          string
	RiderID     string
	DriverID    string
	Pickup      Location
	Dropoff     Location
	Status      RideStatus
	Fare        float64
	Distance    float64
	Duration    time.Duration
	StartTime   *time.Time
	EndTime     *time.Time
	RequestedAt time.Time
	mu          sync.RWMutex
}

// NewRide creates a new ride request
func NewRide(id, riderID string, pickup, dropoff Location) *Ride {
	return &Ride{
		ID:          id,
		RiderID:     riderID,
		Pickup:      pickup,
		Dropoff:     dropoff,
		Status:      RideStatusRequested,
		RequestedAt: time.Now(),
	}
}

// GetStatus returns current ride status
func (r *Ride) GetStatus() RideStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Status
}

// SetStatus updates ride status (thread-safe)
func (r *Ride) SetStatus(status RideStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = status
}

// AssignDriver assigns a driver to the ride
func (r *Ride) AssignDriver(driverID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DriverID = driverID
	r.Status = RideStatusDriverAssigned
}

// Start begins the ride
func (r *Ride) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.StartTime = &now
	r.Status = RideStatusInProgress
}

// Complete ends the ride with fare and distance
func (r *Ride) Complete(fare float64, distance float64, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.EndTime = &now
	r.Fare = fare
	r.Distance = distance
	r.Duration = duration
	r.Status = RideStatusCompleted
}

// Cancel cancels the ride
func (r *Ride) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = RideStatusCancelled
}

// IsInProgress returns true if ride has started
func (r *Ride) IsInProgress() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Status == RideStatusInProgress || r.Status == RideStatusDriverAssigned
}
