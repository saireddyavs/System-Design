package services

import (
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

// SearchCriteria holds search parameters
type SearchCriteria struct {
	Name    string
	Cuisine string
	Location *models.Location
	RadiusKm float64
}

// SearchService handles restaurant search
type SearchService struct {
	restaurantRepo interfaces.RestaurantRepository
}

// NewSearchService creates a new search service
func NewSearchService(restaurantRepo interfaces.RestaurantRepository) *SearchService {
	return &SearchService{restaurantRepo: restaurantRepo}
}

// SearchRestaurants finds restaurants matching the criteria
func (s *SearchService) SearchRestaurants(criteria SearchCriteria) ([]*models.Restaurant, error) {
	var results []*models.Restaurant
	var err error

	// Search by name
	if criteria.Name != "" {
		results, err = s.restaurantRepo.SearchByName(criteria.Name)
		if err != nil {
			return nil, err
		}
	}

	// Filter by cuisine
	if criteria.Cuisine != "" {
		var byCuisine []*models.Restaurant
		byCuisine, err = s.restaurantRepo.SearchByCuisine(criteria.Cuisine)
		if err != nil {
			return nil, err
		}
		if results == nil {
			results = byCuisine
		} else {
			results = intersectRestaurants(results, byCuisine)
		}
	}

	// Filter by location
	if criteria.Location != nil && criteria.RadiusKm > 0 {
		var byLoc []*models.Restaurant
		byLoc, err = s.restaurantRepo.SearchByLocation(*criteria.Location, criteria.RadiusKm)
		if err != nil {
			return nil, err
		}
		if results == nil {
			results = byLoc
		} else {
			results = intersectRestaurants(results, byLoc)
		}
	}

	// If no criteria, return all
	if results == nil {
		results, err = s.restaurantRepo.GetAll()
		if err != nil {
			return nil, err
		}
	}

	return filterOpenRestaurants(results), nil
}

// SearchByCuisine finds restaurants by cuisine type
func (s *SearchService) SearchByCuisine(cuisine string) ([]*models.Restaurant, error) {
	return s.restaurantRepo.SearchByCuisine(cuisine)
}

// SearchByName finds restaurants by name
func (s *SearchService) SearchByName(name string) ([]*models.Restaurant, error) {
	return s.restaurantRepo.SearchByName(name)
}

// SearchByLocation finds restaurants near a location
func (s *SearchService) SearchByLocation(loc models.Location, radiusKm float64) ([]*models.Restaurant, error) {
	return s.restaurantRepo.SearchByLocation(loc, radiusKm)
}

func intersectRestaurants(a, b []*models.Restaurant) []*models.Restaurant {
	seen := make(map[string]bool)
	for _, r := range a {
		seen[r.ID] = true
	}
	var result []*models.Restaurant
	for _, r := range b {
		if seen[r.ID] {
			result = append(result, r)
		}
	}
	return result
}

func filterOpenRestaurants(restaurants []*models.Restaurant) []*models.Restaurant {
	var result []*models.Restaurant
	for _, r := range restaurants {
		if r.IsOpen {
			result = append(result, r)
		}
	}
	return result
}
