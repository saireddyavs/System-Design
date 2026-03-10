package interfaces

import "food-delivery-system/internal/models"

// RestaurantRepository defines the contract for restaurant data access (Repository Pattern)
type RestaurantRepository interface {
	Create(restaurant *models.Restaurant) error
	GetByID(id string) (*models.Restaurant, error)
	GetAll() ([]*models.Restaurant, error)
	Update(restaurant *models.Restaurant) error
	SearchByName(name string) ([]*models.Restaurant, error)
	SearchByCuisine(cuisine string) ([]*models.Restaurant, error)
	SearchByLocation(location models.Location, radiusKm float64) ([]*models.Restaurant, error)
}
