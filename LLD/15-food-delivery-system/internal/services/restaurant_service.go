package services

import (
	"errors"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

var ErrRestaurantNotFound = errors.New("restaurant not found")

// RestaurantService handles restaurant business logic
type RestaurantService struct {
	repo interfaces.RestaurantRepository
}

// NewRestaurantService creates a new restaurant service
func NewRestaurantService(repo interfaces.RestaurantRepository) *RestaurantService {
	return &RestaurantService{repo: repo}
}

// RegisterRestaurant creates a new restaurant
func (s *RestaurantService) RegisterRestaurant(id, name string, cuisines []string, loc models.Location, minOrder float64) (*models.Restaurant, error) {
	restaurant := models.NewRestaurant(id, name, cuisines, loc, minOrder)
	if err := s.repo.Create(restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

// GetRestaurant retrieves a restaurant by ID
func (s *RestaurantService) GetRestaurant(id string) (*models.Restaurant, error) {
	return s.repo.GetByID(id)
}

// AddMenuItem adds a menu item to a restaurant
func (s *RestaurantService) AddMenuItem(restaurantID string, item models.MenuItem) error {
	restaurant, err := s.repo.GetByID(restaurantID)
	if err != nil {
		return err
	}
	item.RestaurantID = restaurantID
	restaurant.AddMenuItem(item)
	return s.repo.Update(restaurant)
}

// UpdateRestaurantStatus sets restaurant open/closed
func (s *RestaurantService) UpdateRestaurantStatus(id string, isOpen bool) error {
	restaurant, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	restaurant.SetOpen(isOpen)
	return s.repo.Update(restaurant)
}

// ListRestaurants returns all restaurants
func (s *RestaurantService) ListRestaurants() ([]*models.Restaurant, error) {
	return s.repo.GetAll()
}
