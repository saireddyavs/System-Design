package tests

import (
	"food-delivery-system/internal/models"
	"food-delivery-system/internal/repositories"
	"food-delivery-system/internal/services"
	"testing"
)

func setupSearchService(t *testing.T) *services.SearchService {
	restaurantRepo := repositories.NewInMemoryRestaurantRepo()
	restaurantService := services.NewRestaurantService(restaurantRepo)

	restaurantService.RegisterRestaurant("R1", "Pizza Paradise", []string{"Italian", "Pizza"}, models.Location{Lat: 12.96, Lng: 77.58}, 100)
	restaurantService.RegisterRestaurant("R2", "Spice Garden", []string{"Indian"}, models.Location{Lat: 12.98, Lng: 77.60}, 150)
	restaurantService.RegisterRestaurant("R3", "Pasta House", []string{"Italian"}, models.Location{Lat: 12.97, Lng: 77.59}, 80)

	return services.NewSearchService(restaurantRepo)
}

func TestSearchByCuisine(t *testing.T) {
	searchService := setupSearchService(t)

	results, err := searchService.SearchByCuisine("Italian")
	if err != nil {
		t.Fatalf("SearchByCuisine failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 Italian restaurants, got %d", len(results))
	}
}

func TestSearchByName(t *testing.T) {
	searchService := setupSearchService(t)

	results, err := searchService.SearchByName("Pizza")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 restaurant with Pizza in name, got %d", len(results))
	}
	if len(results) > 0 && results[0].Name != "Pizza Paradise" {
		t.Errorf("Expected Pizza Paradise, got %s", results[0].Name)
	}
}

func TestSearchByLocation(t *testing.T) {
	searchService := setupSearchService(t)

	loc := models.Location{Lat: 12.97, Lng: 77.59}
	results, err := searchService.SearchByLocation(loc, 5.0)
	if err != nil {
		t.Fatalf("SearchByLocation failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 restaurants within 5km, got %d", len(results))
	}
}

func TestSearchRestaurants_CombinedCriteria(t *testing.T) {
	searchService := setupSearchService(t)

	loc := models.Location{Lat: 12.97, Lng: 77.59}
	results, err := searchService.SearchRestaurants(services.SearchCriteria{
		Name:     "Pizza",
		Cuisine:  "Italian",
		Location: &loc,
		RadiusKm: 10,
	})
	if err != nil {
		t.Fatalf("SearchRestaurants failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 restaurant (Pizza Paradise), got %d", len(results))
	}
	if len(results) > 0 && results[0].Name != "Pizza Paradise" {
		t.Errorf("Expected Pizza Paradise, got %s", results[0].Name)
	}
}
