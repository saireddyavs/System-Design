package models

import "time"

// Ticket represents a parking ticket issued when a vehicle is parked.
// Immutable after creation - used for fee calculation and tracking.
type Ticket struct {
	ID           string
	Vehicle      Vehicle
	SpotID       string
	LevelID      string
	EntryTime    time.Time
	LicensePlate string
}

// NewTicket creates a new parking ticket.
func NewTicket(id string, vehicle Vehicle, spotID, levelID string) *Ticket {
	return &Ticket{
		ID:           id,
		Vehicle:      vehicle,
		SpotID:       spotID,
		LevelID:      levelID,
		EntryTime:    time.Now(),
		LicensePlate: vehicle.GetLicensePlate(),
	}
}

// GetDuration returns how long the vehicle has been parked.
func (t *Ticket) GetDuration() time.Duration {
	return time.Since(t.EntryTime)
}
