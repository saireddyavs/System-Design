package models

import "time"

// User represents a registered user in the system.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose in JSON
	Address   string    `json:"address"`
	CartID    string    `json:"cart_id"`
	CreatedAt time.Time `json:"created_at"`
}
