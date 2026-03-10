package models

import "time"

// Address represents a user's shipping/billing address
type Address struct {
	ID          string `json:"id"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
	IsDefault   bool   `json:"is_default"`
}

// User represents a registered user
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never serialize to JSON
	Addresses []Address `json:"addresses"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
