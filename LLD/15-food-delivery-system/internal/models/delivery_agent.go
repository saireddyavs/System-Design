package models

import (
	"sync"
	"time"
)

// AgentStatus represents the current status of a delivery agent
type AgentStatus string

const (
	AgentStatusAvailable AgentStatus = "available"
	AgentStatusOnDelivery AgentStatus = "on_delivery"
	AgentStatusOffline    AgentStatus = "offline"
)

// DeliveryAgent represents a delivery person
type DeliveryAgent struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Phone     string      `json:"phone"`
	Location  Location    `json:"location"`
	Status    AgentStatus `json:"status"`
	Rating    float64     `json:"rating"`
	CreatedAt time.Time   `json:"created_at"`
	mu        sync.RWMutex
}

// NewDeliveryAgent creates a new delivery agent
func NewDeliveryAgent(id, name, phone string, loc Location) *DeliveryAgent {
	return &DeliveryAgent{
		ID:        id,
		Name:      name,
		Phone:     phone,
		Location:  loc,
		Status:    AgentStatusAvailable,
		Rating:    0,
		CreatedAt: time.Now(),
	}
}

// SetStatus updates the agent's status
func (a *DeliveryAgent) SetStatus(status AgentStatus) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Status = status
}

// UpdateLocation updates the agent's location
func (a *DeliveryAgent) UpdateLocation(loc Location) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Location = loc
}

// UpdateRating updates the agent's rating
func (a *DeliveryAgent) UpdateRating(rating float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Rating = rating
}
