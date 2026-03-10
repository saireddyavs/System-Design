package models

import "time"

// Product represents a sellable item in the catalog
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CategoryID  string    `json:"category_id"`
	Stock       int       `json:"stock"`
	SKU         string    `json:"sku"`
	Weight      float64   `json:"weight"` // in kg
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HasStock checks if product has sufficient quantity
func (p *Product) HasStock(quantity int) bool {
	return p.Stock >= quantity && quantity > 0
}

// DecrementStock reduces stock by quantity (caller must validate)
func (p *Product) DecrementStock(quantity int) {
	if p.Stock >= quantity {
		p.Stock -= quantity
	}
}
