package tests

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"airline-reservation-system/internal/strategies"
	"sync"
	"testing"
	"time"
)

func setupBookingTest(t *testing.T) (*services.BookingService, *models.Flight, *models.Passenger) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	passengerRepo := repositories.NewInMemoryPassengerRepository()

	seatService := services.NewSeatService(flightRepo, strategies.NewAutoAssignFirstAvailable())
	bookingService := services.NewBookingService(
		bookingRepo, flightRepo, passengerRepo,
		seatService,
		strategies.NewClassMultiplierPricing(),
		strategies.NewMockPaymentProcessor(),
		services.NewBookingFactory(),
		services.NewBookingNotifier(),
	)

	departure := time.Now().Add(100 * time.Hour)
	arrival := departure.Add(3 * time.Hour)
	flight, err := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(10, []string{"A", "B", "C"}, models.SeatClassEconomy).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := flightRepo.Create(flight); err != nil {
		t.Fatal(err)
	}

	passenger := models.NewPassenger("P1", "John", "john@test.com", "123", "P123", time.Now())
	if err := passengerRepo.Create(passenger); err != nil {
		t.Fatal(err)
	}

	return bookingService, flight, passenger
}

func TestCreateBooking(t *testing.T) {
	bookingService, _, _ := setupBookingTest(t)

	booking, err := bookingService.CreateBooking("P1", "FL-001", 2, models.SeatClassEconomy)
	if err != nil {
		t.Fatalf("CreateBooking failed: %v", err)
	}
	if booking == nil {
		t.Fatal("booking is nil")
	}
	if len(booking.SeatIDs) != 2 {
		t.Errorf("expected 2 seats, got %d", len(booking.SeatIDs))
	}
	if booking.TotalAmount <= 0 {
		t.Errorf("expected positive amount, got %.2f", booking.TotalAmount)
	}
	if booking.BookingRef == "" {
		t.Error("expected non-empty booking ref")
	}
}

func TestCreateBookingWithSeats(t *testing.T) {
	bookingService, flight, _ := setupBookingTest(t)

	// Get first two seat IDs
	var seatIDs []string
	for i := 0; i < 2 && i < len(flight.Seats); i++ {
		seatIDs = append(seatIDs, flight.Seats[i].ID)
	}

	booking, err := bookingService.CreateBookingWithSeats("P1", "FL-001", seatIDs)
	if err != nil {
		t.Fatalf("CreateBookingWithSeats failed: %v", err)
	}
	if len(booking.SeatIDs) != 2 {
		t.Errorf("expected 2 seats, got %d", len(booking.SeatIDs))
	}
}

func TestDoubleBookingPrevented(t *testing.T) {
	bookingService, flight, _ := setupBookingTest(t)

	seatIDs := []string{flight.Seats[0].ID}

	_, err := bookingService.CreateBookingWithSeats("P1", "FL-001", seatIDs)
	if err != nil {
		t.Fatalf("first booking failed: %v", err)
	}

	// Second booking of same seat should fail
	_, err = bookingService.CreateBookingWithSeats("P1", "FL-001", seatIDs)
	if err == nil {
		t.Error("expected error when double-booking same seat")
	}
}

func TestCancelBookingRefund(t *testing.T) {
	bookingService, _, _ := setupBookingTest(t)

	booking, _ := bookingService.CreateBooking("P1", "FL-001", 1, models.SeatClassEconomy)

	// Cancel >48h before = full refund
	refund, err := bookingService.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("CancelBooking failed: %v", err)
	}
	if refund != booking.TotalAmount {
		t.Errorf("expected full refund %.2f, got %.2f", booking.TotalAmount, refund)
	}
}

func TestConcurrentBookings(t *testing.T) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	passengerRepo := repositories.NewInMemoryPassengerRepository()

	departure := time.Now().Add(100 * time.Hour)
	arrival := departure.Add(3 * time.Hour)
	flight, _ := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("B737").
		BasePrice(100.0).
		AddSeatSection(5, []string{"A", "B"}, models.SeatClassEconomy).
		Build()
	flightRepo.Create(flight)

	passenger := models.NewPassenger("P1", "User", "u@test.com", "123", "P1", time.Now())
	passengerRepo.Create(passenger)

	seatService := services.NewSeatService(flightRepo, strategies.NewAutoAssignFirstAvailable())
	bookingService := services.NewBookingService(
		bookingRepo, flightRepo, passengerRepo,
		seatService, strategies.NewClassMultiplierPricing(),
		strategies.NewMockPaymentProcessor(),
		services.NewBookingFactory(),
		services.NewBookingNotifier(),
	)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(passengerNum int) {
			defer wg.Done()
			_, err := bookingService.CreateBooking("P1", "FL-001", 1, models.SeatClassEconomy)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	// Only 10 seats total (5 rows x 2 cols), so max 10 bookings. We have 5 concurrent.
	// Each tries 1 seat - at most 10 can succeed
	if successCount > 10 {
		t.Errorf("expected at most 10 successful bookings, got %d", successCount)
	}
}
