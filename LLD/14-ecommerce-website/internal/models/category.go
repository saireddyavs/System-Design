package models

import "time"

// Category represents a product category in the catalog
type Category struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ParentID    *string    `json:"parent_id,omitempty"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
