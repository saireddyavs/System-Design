package models

import "time"

// CartItem represents a single item in the shopping cart.
type CartItem struct {
	BookID   string `json:"book_id"`
	Quantity int    `json:"quantity"`
}

// Cart represents a user's shopping cart.
// Items map: BookID -> Quantity
type Cart struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Items     map[string]int `json:"items"` // BookID -> Quantity
	UpdatedAt time.Time `json:"updated_at"`
}
