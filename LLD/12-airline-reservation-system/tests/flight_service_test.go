package tests

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"testing"
	"time"
)

func TestFlightService_AddAndCancelFlight(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	flightService := services.NewFlightService(flightRepo, bookingRepo)

	departure := time.Now().Add(100 * time.Hour)
	arrival := departure.Add(3 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(2, []string{"A", "B"}, models.SeatClassEconomy).
		Build()

	err := flightService.AddFlight(flight)
	if err != nil {
		t.Fatalf("AddFlight failed: %v", err)
	}

	retrieved, err := flightService.GetFlight("FL-001")
	if err != nil {
		t.Fatalf("GetFlight failed: %v", err)
	}
	if retrieved.FlightNumber != "AA100" {
		t.Errorf("expected AA100, got %s", retrieved.FlightNumber)
	}

	err = flightService.CancelFlight("FL-001")
	if err != nil {
		t.Fatalf("CancelFlight failed: %v", err)
	}

	cancelled, _ := flightService.GetFlight("FL-001")
	if cancelled.Status != models.FlightStatusCancelled {
		t.Errorf("expected status Cancelled, got %s", cancelled.Status)
	}
}

func TestFlightService_SearchFlights(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	flightService := services.NewFlightService(flightRepo, bookingRepo)

	departure := time.Date(2025, 7, 20, 10, 0, 0, 0, time.UTC)
	arrival := departure.Add(2 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA200").
		Route("JFK", "SFO").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(200.0).
		AddSeatSection(3, []string{"A"}, models.SeatClassEconomy).
		Build()
	flightRepo.Create(flight)

	results, err := flightService.SearchFlights("JFK", "SFO", time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SearchFlights failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 flight, got %d", len(results))
	}
	if results[0].FlightNumber != "AA200" {
		t.Errorf("expected AA200, got %s", results[0].FlightNumber)
	}
}
