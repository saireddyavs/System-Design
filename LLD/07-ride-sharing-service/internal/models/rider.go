package models

import (
	"sync"
	"time"
)

// Rider represents a ride-sharing customer
type Rider struct {
	ID        string
	Name      string
	Phone     string
	Email     string
	Location  Location
	Rating    float64
	CreatedAt time.Time
	mu        sync.RWMutex
}

// NewRider creates a new rider instance
func NewRider(id, name, phone, email string) *Rider {
	return &Rider{
		ID:        id,
		Name:      name,
		Phone:     phone,
		Email:     email,
		Rating:    5.0,
		CreatedAt: time.Now(),
	}
}

// GetLocation returns a copy of the rider's location
func (r *Rider) GetLocation() Location {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Location
}

// UpdateLocation updates rider's location (thread-safe)
func (r *Rider) UpdateLocation(loc Location) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Location = loc
}

// UpdateRating recalculates average rating
func (r *Rider) UpdateRating(newRating float64, totalRatings int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rating = (r.Rating*float64(totalRatings-1) + newRating) / float64(totalRatings)
}
