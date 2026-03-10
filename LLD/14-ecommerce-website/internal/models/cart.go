package models

import "time"

// CartItem represents a single item in the shopping cart
type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Cart represents a user's shopping cart
type Cart struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Items     map[string]CartItem `json:"items"` // ProductID -> CartItem
	UpdatedAt time.Time          `json:"updated_at"`
}
