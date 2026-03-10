package tests

import (
	"testing"
	"time"

	"hotel-management-system/internal/factory"
	"hotel-management-system/internal/models"
	"hotel-management-system/internal/observer"
	"hotel-management-system/internal/payment"
	"hotel-management-system/internal/repositories"
	"hotel-management-system/internal/services"
	"hotel-management-system/internal/strategies"
)

func setupBookingTest(t *testing.T) (*services.BookingService, *services.RoomService, *services.GuestService) {
	roomRepo := repositories.NewInMemoryRoomRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	guestRepo := repositories.NewInMemoryGuestRepository()
	paymentRepo := repositories.NewInMemoryPaymentRepository()

	roomFactory := factory.NewRoomFactory()
	pricingStrategy := strategies.NewCompositePricingStrategy()
	notificationSvc := observer.NewNotificationService()
	paymentProcessor := payment.NewMockPaymentProcessor()

	roomSvc := services.NewRoomService(roomRepo, bookingRepo, roomFactory)
	guestSvc := services.NewGuestService(guestRepo)
	paymentSvc := services.NewPaymentService(paymentRepo, paymentProcessor)
	bookingSvc := services.NewBookingService(
		bookingRepo, roomRepo, guestRepo, paymentRepo,
		paymentSvc, pricingStrategy, notificationSvc,
	)

	// Seed room and guest
	room, err := roomSvc.CreateRoom("R1", "101", models.RoomTypeDouble, 1)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	_ = room

	guest, err := guestSvc.RegisterGuest("G1", "John", "john@test.com", "123", "PASS1")
	if err != nil {
		t.Fatalf("register guest: %v", err)
	}
	_ = guest

	return bookingSvc, roomSvc, guestSvc
}

func TestBookingFlow_PendingToConfirmedToCheckInToCheckOut(t *testing.T) {
	bookingSvc, _, _ := setupBookingTest(t)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	booking, err := bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	if booking.GetStatus() != models.BookingStatusPending {
		t.Errorf("expected Pending, got %s", booking.GetStatus())
	}

	pmt := models.NewPayment("P1", booking.ID, booking.GetTotalAmount(), models.PaymentMethodCard)
	if err := bookingSvc.ConfirmBooking(booking.ID, pmt); err != nil {
		t.Fatalf("confirm booking: %v", err)
	}

	b, _ := bookingSvc.GetBooking(booking.ID)
	if b.GetStatus() != models.BookingStatusConfirmed {
		t.Errorf("expected Confirmed, got %s", b.GetStatus())
	}

	if err := bookingSvc.CheckIn(booking.ID); err != nil {
		t.Fatalf("check-in: %v", err)
	}
	b, _ = bookingSvc.GetBooking(booking.ID)
	if b.GetStatus() != models.BookingStatusCheckedIn {
		t.Errorf("expected CheckedIn, got %s", b.GetStatus())
	}

	if err := bookingSvc.CheckOut(booking.ID); err != nil {
		t.Fatalf("check-out: %v", err)
	}
	b, _ = bookingSvc.GetBooking(booking.ID)
	if b.GetStatus() != models.BookingStatusCheckedOut {
		t.Errorf("expected CheckedOut, got %s", b.GetStatus())
	}
}

func TestBooking_CancellationRefundPolicy(t *testing.T) {
	bookingSvc, _, _ := setupBookingTest(t)

	// >24h before check-in: full refund
	checkIn := time.Now().Add(48 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 2)

	booking, err := bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	pmt := models.NewPayment("P1", booking.ID, booking.GetTotalAmount(), models.PaymentMethodCard)
	if err := bookingSvc.ConfirmBooking(booking.ID, pmt); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	refund, err := bookingSvc.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	expected := booking.GetTotalAmount()
	if refund < expected*0.99 || refund > expected*1.01 {
		t.Errorf("expected full refund ~%.2f, got %.2f", expected, refund)
	}
}

func TestBooking_OverbookingPrevention(t *testing.T) {
	bookingSvc, _, _ := setupBookingTest(t)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	b1, err := bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
	if err != nil {
		t.Fatalf("create booking 1: %v", err)
	}

	// Same room, overlapping dates - should fail
	_, err = bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
	if err == nil {
		t.Error("expected error for overlapping booking")
	}

	// Confirm first to hold the room
	pmt := models.NewPayment("P1", b1.ID, b1.GetTotalAmount(), models.PaymentMethodCard)
	if err := bookingSvc.ConfirmBooking(b1.ID, pmt); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	// Another overlapping should still fail (room reserved)
	_, err = bookingSvc.CreateBooking("G1", "R1", checkIn.AddDate(0, 0, 1), checkOut.AddDate(0, 0, 1))
	if err == nil {
		t.Error("expected error for overlapping with reserved room")
	}
}

func TestBooking_InvalidStateTransition(t *testing.T) {
	bookingSvc, _, _ := setupBookingTest(t)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	booking, err := bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	// CheckIn without Confirm should fail
	if err := bookingSvc.CheckIn(booking.ID); err == nil {
		t.Error("expected error when check-in without confirm")
	}

	// CheckOut without CheckIn should fail
	pmt := models.NewPayment("P1", booking.ID, booking.GetTotalAmount(), models.PaymentMethodCard)
	if err := bookingSvc.ConfirmBooking(booking.ID, pmt); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if err := bookingSvc.CheckOut(booking.ID); err == nil {
		t.Error("expected error when check-out without check-in")
	}
}
