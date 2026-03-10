package models

import (
	"sync"
	"time"
)

// Customer represents a food delivery customer
type Customer struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Location  Location   `json:"location"`
	Addresses []Address  `json:"addresses"`
	CreatedAt time.Time  `json:"created_at"`
	mu        sync.RWMutex
}

// Address represents a delivery address
type Address struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Location Location `json:"location"`
	Details  string   `json:"details"`
}

// NewCustomer creates a new customer
func NewCustomer(id, name, email, phone string, loc Location) *Customer {
	return &Customer{
		ID:        id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		Location:  loc,
		Addresses: []Address{},
		CreatedAt: time.Now(),
	}
}

// AddAddress adds a delivery address
func (c *Customer) AddAddress(addr Address) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Addresses = append(c.Addresses, addr)
}

// UpdateLocation updates the customer's current location
func (c *Customer) UpdateLocation(loc Location) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Location = loc
}
