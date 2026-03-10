package models

// MenuItem represents a food item in a restaurant's menu
type MenuItem struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
	IsAvailable  bool    `json:"is_available"`
}

// NewMenuItem creates a new menu item
func NewMenuItem(id, restaurantID, name, description, category string, price float64) MenuItem {
	return MenuItem{
		ID:           id,
		RestaurantID: restaurantID,
		Name:         name,
		Description:  description,
		Price:        price,
		Category:     category,
		IsAvailable:  true,
	}
}
