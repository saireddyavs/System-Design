package models

import (
	"sync"
	"time"
)

// DriverStatus represents the availability state of a driver
type DriverStatus string

const (
	DriverStatusAvailable DriverStatus = "available"
	DriverStatusOnRide    DriverStatus = "on_ride"
	DriverStatusOffline   DriverStatus = "offline"
	DriverStatusDeactivated DriverStatus = "deactivated"
)

// Vehicle represents driver's vehicle information
type Vehicle struct {
	Model  string
	Number string
	Type   string
}

// Driver represents a ride-sharing driver
type Driver struct {
	ID        string
	Name      string
	Phone     string
	Vehicle   Vehicle
	Location  Location
	Status    DriverStatus
	Rating    float64
	CreatedAt time.Time
	mu        sync.RWMutex
}

// NewDriver creates a new driver instance
func NewDriver(id, name, phone string, vehicle Vehicle) *Driver {
	return &Driver{
		ID:        id,
		Name:      name,
		Phone:     phone,
		Vehicle:   vehicle,
		Status:    DriverStatusOffline,
		Rating:    5.0,
		CreatedAt: time.Now(),
	}
}

// IsAvailable returns true if driver can accept rides
func (d *Driver) IsAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status == DriverStatusAvailable
}

// GetLocation returns a copy of the driver's location
func (d *Driver) GetLocation() Location {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Location
}

// UpdateLocation updates driver's location (thread-safe)
func (d *Driver) UpdateLocation(loc Location) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Location = loc
}

// SetStatus updates driver status (thread-safe)
func (d *Driver) SetStatus(status DriverStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = status
}

// GetStatus returns current driver status
func (d *Driver) GetStatus() DriverStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status
}

// GetRating returns current driver rating (thread-safe)
func (d *Driver) GetRating() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Rating
}

// UpdateRating recalculates average rating
func (d *Driver) UpdateRating(newRating float64, totalRatings int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Rating = (d.Rating*float64(totalRatings-1) + newRating) / float64(totalRatings)
}
