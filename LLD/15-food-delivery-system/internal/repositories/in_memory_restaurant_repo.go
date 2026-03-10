package repositories

import (
	"errors"
	"food-delivery-system/internal/models"
	"strings"
	"sync"
)

var ErrRestaurantNotFound = errors.New("restaurant not found")

// InMemoryRestaurantRepo implements RestaurantRepository with thread-safe in-memory storage
type InMemoryRestaurantRepo struct {
	restaurants map[string]*models.Restaurant
	mu          sync.RWMutex
}

// NewInMemoryRestaurantRepo creates a new in-memory restaurant repository
func NewInMemoryRestaurantRepo() *InMemoryRestaurantRepo {
	return &InMemoryRestaurantRepo{
		restaurants: make(map[string]*models.Restaurant),
	}
}

func (r *InMemoryRestaurantRepo) Create(restaurant *models.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.restaurants[restaurant.ID]; exists {
		return errors.New("restaurant already exists")
	}
	r.restaurants[restaurant.ID] = restaurant
	return nil
}

func (r *InMemoryRestaurantRepo) GetByID(id string) (*models.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	restaurant, exists := r.restaurants[id]
	if !exists {
		return nil, ErrRestaurantNotFound
	}
	return restaurant, nil
}

func (r *InMemoryRestaurantRepo) GetAll() ([]*models.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Restaurant, 0, len(r.restaurants))
	for _, r := range r.restaurants {
		result = append(result, r)
	}
	return result, nil
}

func (r *InMemoryRestaurantRepo) Update(restaurant *models.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.restaurants[restaurant.ID]; !exists {
		return ErrRestaurantNotFound
	}
	r.restaurants[restaurant.ID] = restaurant
	return nil
}

func (r *InMemoryRestaurantRepo) SearchByName(name string) ([]*models.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name = strings.ToLower(name)
	var result []*models.Restaurant
	for _, rest := range r.restaurants {
		if strings.Contains(strings.ToLower(rest.Name), name) {
			result = append(result, rest)
		}
	}
	return result, nil
}

func (r *InMemoryRestaurantRepo) SearchByCuisine(cuisine string) ([]*models.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cuisine = strings.ToLower(cuisine)
	var result []*models.Restaurant
	for _, rest := range r.restaurants {
		for _, c := range rest.Cuisines {
			if strings.ToLower(c) == cuisine {
				result = append(result, rest)
				break
			}
		}
	}
	return result, nil
}

func (r *InMemoryRestaurantRepo) SearchByLocation(location models.Location, radiusKm float64) ([]*models.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Restaurant
	for _, rest := range r.restaurants {
		if rest.IsOpen && location.Distance(rest.Location) <= radiusKm {
			result = append(result, rest)
		}
	}
	return result, nil
}
