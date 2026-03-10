package models

import "time"

// Cart represents a user's shopping cart with items and status
type Cart struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Items     []CartItem  `json:"items"`
	Status    CartStatus  `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Subtotal calculates sum of all item subtotals
func (c *Cart) Subtotal() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Subtotal
	}
	return total
}

// ItemCount returns total number of items (sum of quantities)
func (c *Cart) ItemCount() int {
	var count int
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

// IsEmpty returns true if cart has no items
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}
