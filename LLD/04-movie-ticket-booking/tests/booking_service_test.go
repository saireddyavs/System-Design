package tests

import (
	"movie-ticket-booking/internal/models"
	"movie-ticket-booking/internal/repositories"
	"movie-ticket-booking/internal/services"
	"movie-ticket-booking/internal/strategies"
	"sync"
	"testing"
	"time"
)

func setupBookingTest(t *testing.T) *services.BookingService {
	screenRepo := repositories.NewInMemoryScreenRepository()
	showRepo := repositories.NewInMemoryShowRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	userRepo.Create(&models.User{ID: "u1", Name: "Test", Email: "test@test.com", Phone: "123"})
	screen := &models.Screen{
		ID: "s1", TheatreID: "t1", Name: "Screen 1", TotalCapacity: 4,
		Seats: []models.Seat{
			{ID: "seat-1", ScreenID: "s1", Row: "A", Number: 1, Category: models.SeatCategoryRegular},
			{ID: "seat-2", ScreenID: "s1", Row: "A", Number: 2, Category: models.SeatCategoryRegular},
		},
	}
	screenRepo.Create(screen)

	start := time.Now().Add(48 * time.Hour)
	show := &models.Show{
		ID: "show-1", MovieID: "m1", ScreenID: "s1", TheatreID: "t1",
		StartTime: start, EndTime: start.Add(120 * time.Minute),
		SeatStatusMap: map[string]models.SeatStatus{"seat-1": models.SeatStatusAvailable, "seat-2": models.SeatStatusAvailable},
		BasePrice:     100,
	}
	showRepo.Create(show)

	payment := services.NewMockPaymentProcessor()
	notification := services.NewObserverNotificationService()
	pricing := strategies.NewWeekdayPricingStrategy()

	return services.NewBookingService(
		bookingRepo, showRepo, screenRepo, userRepo,
		payment, notification, pricing,
	)
}

func TestCreateBooking_Success(t *testing.T) {
	svc := setupBookingTest(t)

	booking, err := svc.CreateBooking(&services.CreateBookingRequest{
		UserID:  "u1",
		ShowID:  "show-1",
		SeatIDs: []string{"seat-1", "seat-2"},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking == nil {
		t.Fatal("expected booking, got nil")
	}
	if booking.Status != models.BookingStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", booking.Status)
	}
	if len(booking.SeatIDs) != 2 {
		t.Errorf("expected 2 seats, got %d", len(booking.SeatIDs))
	}
	if booking.TotalAmount <= 0 {
		t.Errorf("expected positive amount, got %.2f", booking.TotalAmount)
	}
}

func TestCreateBooking_SameSeatTwice_Fails(t *testing.T) {
	svc := setupBookingTest(t)

	b1, err := svc.CreateBooking(&services.CreateBookingRequest{
		UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-1"},
	})
	if err != nil {
		t.Fatalf("first booking failed: %v", err)
	}
	if b1 == nil {
		t.Fatal("first booking nil")
	}

	_, err = svc.CreateBooking(&services.CreateBookingRequest{
		UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-1"},
	})
	if err == nil {
		t.Fatal("expected error for double booking, got nil")
	}
	if err != services.ErrSeatNotAvailable {
		t.Errorf("expected ErrSeatNotAvailable, got %v", err)
	}
}

func TestCancelBooking_FullRefund(t *testing.T) {
	svc := setupBookingTest(t)

	booking, _ := svc.CreateBooking(&services.CreateBookingRequest{
		UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-1"},
	})

	cancelled, err := svc.CancelBooking(booking.ID)
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if cancelled.Status != models.BookingStatusCancelled {
		t.Errorf("expected cancelled status, got %s", cancelled.Status)
	}
	if cancelled.RefundAmount != booking.TotalAmount {
		t.Errorf("expected full refund %.2f, got %.2f", booking.TotalAmount, cancelled.RefundAmount)
	}
}

func TestConcurrentBooking_SameSeat_OnlyOneSucceeds(t *testing.T) {
	svc := setupBookingTest(t)

	var wg sync.WaitGroup
	results := make([]*models.Booking, 10)
	errs := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			b, err := svc.CreateBooking(&services.CreateBookingRequest{
				UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-1"},
			})
			results[idx] = b
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	successCount := 0
	for i := 0; i < 10; i++ {
		if errs[i] == nil && results[i] != nil {
			successCount++
		}
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful booking, got %d", successCount)
	}
}

// TestConcurrentBooking_DifferentSeats_BothSucceed verifies concurrent bookings for different seats work
func TestConcurrentBooking_DifferentSeats_BothSucceed(t *testing.T) {
	svc := setupBookingTest(t)

	var wg sync.WaitGroup
	var b1, b2 *models.Booking
	var err1, err2 error

	wg.Add(2)
	go func() {
		defer wg.Done()
		b1, err1 = svc.CreateBooking(&services.CreateBookingRequest{
			UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-1"},
		})
	}()
	go func() {
		defer wg.Done()
		b2, err2 = svc.CreateBooking(&services.CreateBookingRequest{
			UserID: "u1", ShowID: "show-1", SeatIDs: []string{"seat-2"},
		})
	}()
	wg.Wait()

	if err1 != nil || b1 == nil {
		t.Errorf("booking 1 failed: err=%v", err1)
	}
	if err2 != nil || b2 == nil {
		t.Errorf("booking 2 failed: err=%v", err2)
	}
	if b1 != nil && b2 != nil && b1.ID == b2.ID {
		t.Error("expected different booking IDs")
	}
}
