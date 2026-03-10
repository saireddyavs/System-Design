package tests

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"airline-reservation-system/internal/strategies"
	"testing"
	"time"
)

func setupCancellationTest(t *testing.T, hoursUntilDeparture float64) (*services.BookingService, *models.Booking) {
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	passengerRepo := repositories.NewInMemoryPassengerRepository()

	departure := time.Now().Add(time.Duration(hoursUntilDeparture) * time.Hour)
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

	passenger := models.NewPassenger("P1", "John", "j@test.com", "123", "P1", time.Now())
	passengerRepo.Create(passenger)

	seatService := services.NewSeatService(flightRepo, strategies.NewAutoAssignFirstAvailable())
	bookingService := services.NewBookingService(
		bookingRepo, flightRepo, passengerRepo,
		seatService, strategies.NewClassMultiplierPricing(),
		strategies.NewMockPaymentProcessor(),
		services.NewBookingFactory(),
		services.NewBookingNotifier(),
	)

	booking, err := bookingService.CreateBooking("P1", "FL-001", 1, models.SeatClassEconomy)
	if err != nil {
		t.Fatal(err)
	}
	return bookingService, booking
}

func TestCancellationRefund_FullRefund(t *testing.T) {
	bookingService, booking := setupCancellationTest(t, 72) // 72 hours before

	refund, err := bookingService.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("CancelBooking failed: %v", err)
	}
	if refund != booking.TotalAmount {
		t.Errorf("expected full refund %.2f, got %.2f", booking.TotalAmount, refund)
	}
}

func TestCancellationRefund_75Percent(t *testing.T) {
	bookingService, booking := setupCancellationTest(t, 36) // 36 hours before

	refund, err := bookingService.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("CancelBooking failed: %v", err)
	}
	expected := booking.TotalAmount * 0.75
	if refund != expected {
		t.Errorf("expected 75%% refund %.2f, got %.2f", expected, refund)
	}
}

func TestCancellationRefund_25Percent(t *testing.T) {
	bookingService, booking := setupCancellationTest(t, 12) // 12 hours before

	refund, err := bookingService.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("CancelBooking failed: %v", err)
	}
	expected := booking.TotalAmount * 0.25
	if refund != expected {
		t.Errorf("expected 25%% refund %.2f, got %.2f", expected, refund)
	}
}
