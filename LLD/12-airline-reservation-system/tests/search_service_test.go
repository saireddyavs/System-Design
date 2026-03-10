package tests

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"testing"
	"time"
)

func setupSearchTest(t *testing.T) (*services.SearchService, *models.Flight) {
	flightRepo := repositories.NewInMemoryFlightRepository()

	departure := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	arrival := departure.Add(3 * time.Hour)
	flight, err := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(150.0).
		AddSeatSection(5, []string{"A", "B", "C"}, models.SeatClassEconomy).
		AddSeatSection(2, []string{"A", "B"}, models.SeatClassBusiness).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	flightRepo.Create(flight)

	searchService := services.NewSearchService(flightRepo)
	return searchService, flight
}

func TestSearchFlightsByRoute(t *testing.T) {
	searchService, _ := setupSearchTest(t)

	results, err := searchService.SearchFlights(services.SearchCriteria{
		Origin:      "JFK",
		Destination: "LAX",
		Date:        time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("SearchFlights failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Flight.Origin != "JFK" || results[0].Flight.Destination != "LAX" {
		t.Error("wrong flight returned")
	}
	if results[0].AvailableSeats != 19 { // 5*3 economy + 2*2 business
		t.Errorf("expected 19 available seats, got %d", results[0].AvailableSeats)
	}
}

func TestSearchFlightsByClass(t *testing.T) {
	searchService, _ := setupSearchTest(t)

	results, err := searchService.SearchFlights(services.SearchCriteria{
		Origin:      "JFK",
		Destination: "LAX",
		Date:        time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		Class:       models.SeatClassBusiness,
	})
	if err != nil {
		t.Fatalf("SearchFlights failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AvailableSeats != 4 { // 2 rows x 2 cols
		t.Errorf("expected 4 business seats, got %d", results[0].AvailableSeats)
	}
}

func TestSearchNoResults(t *testing.T) {
	searchService, _ := setupSearchTest(t)

	results, err := searchService.SearchFlights(services.SearchCriteria{
		Origin:      "LAX",
		Destination: "JFK",
		Date:        time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("SearchFlights failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for wrong route, got %d", len(results))
	}
}

func TestSearchExcludesCancelledFlights(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()

	departure := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	arrival := departure.Add(3 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(2, []string{"A"}, models.SeatClassEconomy).
		Build()
	flight.Status = models.FlightStatusCancelled
	flightRepo.Create(flight)

	searchService := services.NewSearchService(flightRepo)
	results, _ := searchService.SearchFlights(services.SearchCriteria{
		Origin: "JFK", Destination: "LAX",
		Date: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
	})
	if len(results) != 0 {
		t.Errorf("cancelled flights should be excluded, got %d results", len(results))
	}
}
