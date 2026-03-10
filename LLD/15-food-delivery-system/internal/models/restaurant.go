package models

import (
	"sync"
	"time"
)

// Restaurant represents a food establishment
type Restaurant struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Cuisines  []string   `json:"cuisines"`
	Location  Location   `json:"location"`
	Rating    float64    `json:"rating"`
	IsOpen    bool       `json:"is_open"`
	Menu      []MenuItem `json:"menu"`
	MinOrder  float64    `json:"min_order"`
	CreatedAt time.Time  `json:"created_at"`
	mu        sync.RWMutex
}

// NewRestaurant creates a new restaurant
func NewRestaurant(id, name string, cuisines []string, loc Location, minOrder float64) *Restaurant {
	return &Restaurant{
		ID:        id,
		Name:      name,
		Cuisines:  cuisines,
		Location:  loc,
		Rating:    0,
		IsOpen:    true,
		Menu:      []MenuItem{},
		MinOrder:  minOrder,
		CreatedAt: time.Now(),
	}
}

// AddMenuItem adds a menu item to the restaurant
func (r *Restaurant) AddMenuItem(item MenuItem) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Menu = append(r.Menu, item)
}

// GetMenu returns a copy of the menu
func (r *Restaurant) GetMenu() []MenuItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	menu := make([]MenuItem, len(r.Menu))
	copy(menu, r.Menu)
	return menu
}

// SetOpen sets the restaurant open/closed status
func (r *Restaurant) SetOpen(open bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.IsOpen = open
}

// UpdateRating updates the restaurant rating
func (r *Restaurant) UpdateRating(rating float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rating = rating
}
