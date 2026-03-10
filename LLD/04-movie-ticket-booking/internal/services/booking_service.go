package services

import (
	"errors"
	"fmt"
	"movie-ticket-booking/internal/interfaces"
	"movie-ticket-booking/internal/models"
	"sync/atomic"
	"time"
)

var (
	ErrSeatNotAvailable = errors.New("seat not available")
	ErrShowNotFound     = errors.New("show not found")
	ErrScreenNotFound   = errors.New("screen not found")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrInvalidCancellation = errors.New("booking cannot be cancelled")
)

// BookingService handles ticket booking with concurrency safety
type BookingService struct {
	bookingRepo   interfaces.BookingRepository
	showRepo      interfaces.ShowRepository
	screenRepo    interfaces.ScreenRepository
	userRepo      interfaces.UserRepository
	payment       interfaces.PaymentProcessor
	notification  interfaces.NotificationService
	pricing       interfaces.PricingStrategy
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo interfaces.BookingRepository,
	showRepo interfaces.ShowRepository,
	screenRepo interfaces.ScreenRepository,
	userRepo interfaces.UserRepository,
	payment interfaces.PaymentProcessor,
	notification interfaces.NotificationService,
	pricing interfaces.PricingStrategy,
) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		showRepo:     showRepo,
		screenRepo:   screenRepo,
		userRepo:     userRepo,
		payment:      payment,
		notification: notification,
		pricing:      pricing,
	}
}

// CreateBookingRequest represents booking request
type CreateBookingRequest struct {
	UserID   string
	ShowID   string
	SeatIDs  []string
}

// CreateBooking creates a booking with concurrent seat locking (Factory + Pessimistic Locking)
func (s *BookingService) CreateBooking(req *CreateBookingRequest) (*models.Booking, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, err
	}

	show, err := s.showRepo.GetByID(req.ShowID)
	if err != nil {
		return nil, ErrShowNotFound
	}

	screen, err := s.screenRepo.GetByID(show.ScreenID)
	if err != nil {
		return nil, ErrScreenNotFound
	}

	seatMap := make(map[string]*models.Seat)
	for i := range screen.Seats {
		seatMap[screen.Seats[i].ID] = &screen.Seats[i]
	}

	var totalAmount float64
	var booking *models.Booking

	// Use UpdateSeats for atomic seat allocation - prevents double booking
	err = s.showRepo.UpdateSeats(req.ShowID, func(sh *models.Show) error {
		// Verify all seats available and calculate price
		for _, seatID := range req.SeatIDs {
			seat, ok := seatMap[seatID]
			if !ok {
				return errors.New("invalid seat ID: " + seatID)
			}
			status, exists := sh.SeatStatusMap[seatID]
			if !exists || status != models.SeatStatusAvailable {
				return ErrSeatNotAvailable
			}
			totalAmount += s.pricing.CalculatePrice(sh.BasePrice, seat.Category, sh.StartTime)
		}

		// Mark seats as booked
		for _, seatID := range req.SeatIDs {
			sh.SeatStatusMap[seatID] = models.SeatStatusBooked
		}

		// Create booking (Factory pattern - building booking object)
		booking = s.createBooking(req.UserID, req.ShowID, req.SeatIDs, totalAmount)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		// Rollback seats on failure
		_ = s.showRepo.UpdateSeats(req.ShowID, func(sh *models.Show) error {
			for _, seatID := range req.SeatIDs {
				sh.SeatStatusMap[seatID] = models.SeatStatusAvailable
			}
			return nil
		})
		return nil, err
	}

	if err := s.payment.ProcessPayment(totalAmount, booking.ID); err != nil {
		// Rollback on payment failure
		_ = s.showRepo.UpdateSeats(req.ShowID, func(sh *models.Show) error {
			for _, seatID := range req.SeatIDs {
				sh.SeatStatusMap[seatID] = models.SeatStatusAvailable
			}
			return nil
		})
		booking.Status = models.BookingStatusCancelled
		_ = s.bookingRepo.Update(booking)
		return nil, err
	}

	// Observer: Notify on booking confirmation
	user, _ := s.userRepo.GetByID(req.UserID)
	_ = s.notification.Notify(&interfaces.NotificationPayload{
		Event:   interfaces.EventBookingConfirmed,
		Booking: booking,
		User:    user,
		Message: "Booking confirmed for " + booking.ID,
	})

	return booking, nil
}

// createBooking is the Factory for creating booking objects
func (s *BookingService) createBooking(userID, showID string, seatIDs []string, totalAmount float64) *models.Booking {
	return &models.Booking{
		ID:          generateBookingID(),
		UserID:      userID,
		ShowID:      showID,
		SeatIDs:     seatIDs,
		TotalAmount: totalAmount,
		Status:      models.BookingStatusConfirmed,
		BookedAt:    time.Now(),
	}
}

// CancelBooking cancels a booking with refund policy
func (s *BookingService) CancelBooking(bookingID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	if booking.Status == models.BookingStatusCancelled {
		return nil, ErrInvalidCancellation
	}

	// Refund policy: full refund if >24h before show, 50% otherwise
	show, err := s.showRepo.GetByID(booking.ShowID)
	if err != nil {
		return nil, ErrShowNotFound
	}

	refundAmount := s.calculateRefund(booking.TotalAmount, show.StartTime)

	// Release seats atomically
	err = s.showRepo.UpdateSeats(booking.ShowID, func(sh *models.Show) error {
		for _, seatID := range booking.SeatIDs {
			sh.SeatStatusMap[seatID] = models.SeatStatusAvailable
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	booking.Status = models.BookingStatusCancelled
	booking.CancelledAt = &now
	booking.RefundAmount = refundAmount

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, err
	}

	if err := s.payment.ProcessRefund(refundAmount, bookingID); err != nil {
		// Log but don't fail - seats already released
	}

	// Observer: Notify on cancellation
	user, _ := s.userRepo.GetByID(booking.UserID)
	_ = s.notification.Notify(&interfaces.NotificationPayload{
		Event:   interfaces.EventBookingCancelled,
		Booking: booking,
		User:    user,
		Message: fmt.Sprintf("Booking cancelled. Refund: %.2f", refundAmount),
	})

	return booking, nil
}

func (s *BookingService) calculateRefund(amount float64, showTime time.Time) float64 {
	hoursUntilShow := time.Until(showTime).Hours()
	if hoursUntilShow >= 24 {
		return amount // Full refund
	}
	return amount * 0.5 // 50% refund
}

var bookingIDCounter uint64

func generateBookingID() string {
	return fmt.Sprintf("BKG-%d-%d", time.Now().Unix(), atomic.AddUint64(&bookingIDCounter, 1))
}
