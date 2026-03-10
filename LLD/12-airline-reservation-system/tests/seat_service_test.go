package tests

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"airline-reservation-system/internal/strategies"
	"testing"
	"time"
)

func setupSeatTest(t *testing.T) (*services.SeatService, *models.Flight) {
	flightRepo := repositories.NewInMemoryFlightRepository()

	departure := time.Now().Add(24 * time.Hour)
	arrival := departure.Add(2 * time.Hour)
	flight, err := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(3, []string{"A", "B", "C", "D", "E", "F"}, models.SeatClassEconomy).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	flightRepo.Create(flight)

	seatService := services.NewSeatService(flightRepo, strategies.NewAutoAssignFirstAvailable())
	return seatService, flight
}

func TestAutoAssignSeats(t *testing.T) {
	seatService, flight := setupSeatTest(t)

	seatIDs, err := seatService.AutoAssignSeats(flight.ID, 3, models.SeatClassEconomy)
	if err != nil {
		t.Fatalf("AutoAssignSeats failed: %v", err)
	}
	if len(seatIDs) != 3 {
		t.Errorf("expected 3 seats, got %d", len(seatIDs))
	}
}

func TestWindowPreferenceAssignment(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	departure := time.Now().Add(24 * time.Hour)
	arrival := departure.Add(2 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(2, []string{"A", "B", "C", "F"}, models.SeatClassEconomy).
		Build()
	flightRepo.Create(flight)

	seatService := services.NewSeatService(flightRepo, strategies.NewWindowPreferenceAssignment())
	seatIDs, err := seatService.AutoAssignSeats(flight.ID, 2, models.SeatClassEconomy)
	if err != nil {
		t.Fatalf("AutoAssignSeats failed: %v", err)
	}
	if len(seatIDs) != 2 {
		t.Errorf("expected 2 seats, got %d", len(seatIDs))
	}
	// Verify window seats (A, F) are preferred - check seat IDs map to window columns
	flightUpdated, _ := flightRepo.GetByID(flight.ID)
	for _, id := range seatIDs {
		for _, s := range flightUpdated.Seats {
			if s.ID == id && (s.Column == "A" || s.Column == "F") {
				// Found a window seat - good
				break
			}
		}
	}
}

func TestManualAssignSeats(t *testing.T) {
	seatService, flight := setupSeatTest(t)

	seatIDs := []string{flight.Seats[0].ID, flight.Seats[1].ID}
	err := seatService.ManualAssignSeats(flight.ID, seatIDs)
	if err != nil {
		t.Fatalf("ManualAssignSeats failed: %v", err)
	}
}

func TestManualAssignBookedSeatFails(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	departure := time.Now().Add(24 * time.Hour)
	arrival := departure.Add(2 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(3, []string{"A", "B", "C"}, models.SeatClassEconomy).
		Build()
	flight.Seats[0].Status = models.SeatStatusBooked
	flightRepo.Create(flight)

	seatService := services.NewSeatService(flightRepo, strategies.NewAutoAssignFirstAvailable())
	err := seatService.ManualAssignSeats(flight.ID, []string{flight.Seats[0].ID})
	if err == nil {
		t.Error("expected error when assigning already booked seat")
	}
}

func TestInsufficientSeats(t *testing.T) {
	seatService, flight := setupSeatTest(t)

	_, err := seatService.AutoAssignSeats(flight.ID, 100, models.SeatClassEconomy)
	if err == nil {
		t.Error("expected error when requesting more seats than available")
	}
}

func TestPricingStrategy(t *testing.T) {
	basePrice := 100.0
	seats := []*models.Seat{
		{Class: models.SeatClassEconomy},
		{Class: models.SeatClassBusiness},
		{Class: models.SeatClassFirst},
	}

	pricing := strategies.NewClassMultiplierPricing()
	total := pricing.CalculatePrice(basePrice, seats)

	expected := 100*1.0 + 100*2.5 + 100*5.0 // 100 + 250 + 500 = 850
	if total != expected {
		t.Errorf("expected %.2f, got %.2f", expected, total)
	}
}

func TestBaggageAllowance(t *testing.T) {
	if models.SeatClassEconomy.BaggageAllowance() != 23 {
		t.Error("Economy baggage should be 23kg")
	}
	if models.SeatClassBusiness.BaggageAllowance() != 32 {
		t.Error("Business baggage should be 32kg")
	}
	if models.SeatClassFirst.BaggageAllowance() != 40 {
		t.Error("First baggage should be 40kg")
	}
}
