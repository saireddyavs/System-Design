package services

import (
	"errors"
	"fmt"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
	"time"
)

// RefundPolicy: >24h full, <24h 50%, no-show 0%
const (
	RefundFullHours    = 24
	RefundPartialPct   = 0.5
)

// BookingService orchestrates booking lifecycle (State pattern for status)
type BookingService struct {
	bookingRepo   interfaces.BookingRepository
	roomRepo      interfaces.RoomRepository
	guestRepo     interfaces.GuestRepository
	paymentRepo   interfaces.PaymentRepository
	paymentSvc    *PaymentService
	pricing       interfaces.PricingStrategy
	notification  interfaces.NotificationService
	mu            sync.RWMutex // For overbooking prevention
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo interfaces.BookingRepository,
	roomRepo interfaces.RoomRepository,
	guestRepo interfaces.GuestRepository,
	paymentRepo interfaces.PaymentRepository,
	paymentSvc *PaymentService,
	pricing interfaces.PricingStrategy,
	notification interfaces.NotificationService,
) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		roomRepo:     roomRepo,
		guestRepo:    guestRepo,
		paymentRepo:  paymentRepo,
		paymentSvc:   paymentSvc,
		pricing:      pricing,
		notification: notification,
	}
}

// CreateBooking creates a new booking (Pending state)
func (s *BookingService) CreateBooking(guestID, roomID string, checkIn, checkOut time.Time) (*models.Booking, error) {
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		return nil, errors.New("check-out must be after check-in")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room, err := s.roomRepo.GetByID(roomID)
	if err != nil || room == nil {
		return nil, ErrRoomNotAvailable
	}

	guest, err := s.guestRepo.GetByID(guestID)
	if err != nil || guest == nil {
		return nil, ErrGuestNotFound
	}

	overlapping, err := s.bookingRepo.GetBookingsForRoomInRange(roomID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if len(overlapping) > 0 {
		return nil, ErrRoomNotAvailable
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	ctx := &interfaces.PricingContext{
		Room:         room,
		Guest:        guest,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		Nights:       nights,
	}
	totalAmount := s.pricing.CalculatePrice(ctx)

	booking := models.NewBooking(
		fmt.Sprintf("BKG-%d", time.Now().UnixNano()),
		guestID,
		roomID,
		checkIn,
		checkOut,
		totalAmount,
	)

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// ConfirmBooking moves Pending -> Confirmed (requires payment)
func (s *BookingService) ConfirmBooking(bookingID string, payment *models.Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil || booking == nil {
		return ErrBookingNotFound
	}
	if booking.GetStatus() != models.BookingStatusPending {
		return ErrInvalidStateTransition
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return err
	}
	if err := s.paymentSvc.ProcessPayment(payment); err != nil {
		return err
	}

	booking.SetStatus(models.BookingStatusConfirmed)
	booking.SetPaymentStatus(models.PaymentStatusCompleted)
	if err := s.bookingRepo.Update(booking); err != nil {
		return err
	}

	room, _ := s.roomRepo.GetByID(booking.RoomID)
	room.SetStatus(models.RoomStatusReserved)
	s.roomRepo.Update(room)

	guest, _ := s.guestRepo.GetByID(booking.GuestID)
	s.notification.Notify(interfaces.NotificationPayload{
		Event:   interfaces.EventBookingConfirmed,
		Booking: booking,
		Guest:   guest,
		Room:    room,
	})
	return nil
}

// CheckIn moves Confirmed -> CheckedIn
func (s *BookingService) CheckIn(bookingID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil || booking == nil {
		return ErrBookingNotFound
	}
	if booking.GetStatus() != models.BookingStatusConfirmed {
		return ErrInvalidStateTransition
	}

	booking.SetStatus(models.BookingStatusCheckedIn)
	if err := s.bookingRepo.Update(booking); err != nil {
		return err
	}

	room, _ := s.roomRepo.GetByID(booking.RoomID)
	room.SetStatus(models.RoomStatusOccupied)
	s.roomRepo.Update(room)

	guest, _ := s.guestRepo.GetByID(booking.GuestID)
	s.notification.Notify(interfaces.NotificationPayload{
		Event:   interfaces.EventCheckIn,
		Booking: booking,
		Guest:   guest,
		Room:    room,
	})
	return nil
}

// CheckOut moves CheckedIn -> CheckedOut, awards loyalty points
func (s *BookingService) CheckOut(bookingID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil || booking == nil {
		return ErrBookingNotFound
	}
	if booking.GetStatus() != models.BookingStatusCheckedIn {
		return ErrInvalidStateTransition
	}

	booking.SetStatus(models.BookingStatusCheckedOut)
	if err := s.bookingRepo.Update(booking); err != nil {
		return err
	}

	room, _ := s.roomRepo.GetByID(booking.RoomID)
	room.SetStatus(models.RoomStatusAvailable)
	s.roomRepo.Update(room)

	guest, _ := s.guestRepo.GetByID(booking.GuestID)
	points := booking.Nights() * 100
	guest.AddLoyaltyPoints(points)
	s.guestRepo.Update(guest)

	s.notification.Notify(interfaces.NotificationPayload{
		Event:   interfaces.EventCheckOut,
		Booking: booking,
		Guest:   guest,
		Room:    room,
	})
	return nil
}

// CancelBooking moves to Cancelled, applies refund policy
func (s *BookingService) CancelBooking(bookingID string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil || booking == nil {
		return 0, ErrBookingNotFound
	}
	status := booking.GetStatus()
	if status != models.BookingStatusPending && status != models.BookingStatusConfirmed {
		return 0, ErrInvalidStateTransition
	}

	hoursUntilCheckIn := time.Until(booking.GetCheckInDate()).Hours()
	var refundPct float64
	if hoursUntilCheckIn >= RefundFullHours {
		refundPct = 1.0
	} else if hoursUntilCheckIn > 0 {
		refundPct = RefundPartialPct
	} else {
		refundPct = 0 // No-show
	}

	payments, _ := s.paymentRepo.GetByBookingID(bookingID)
	refundAmount := 0.0
	for _, p := range payments {
		if p.GetStatus() == models.PaymentStatusCompleted {
			amt := p.GetAmount() * refundPct
			refundAmount += amt
			if amt > 0 {
				_ = s.paymentSvc.ProcessRefund(p, amt)
			}
		}
	}

	booking.SetStatus(models.BookingStatusCancelled)
	s.bookingRepo.Update(booking)

	if status == models.BookingStatusConfirmed {
		room, _ := s.roomRepo.GetByID(booking.RoomID)
		room.SetStatus(models.RoomStatusAvailable)
		s.roomRepo.Update(room)
	}

	guest, _ := s.guestRepo.GetByID(booking.GuestID)
	s.notification.Notify(interfaces.NotificationPayload{
		Event:   interfaces.EventBookingCancelled,
		Booking: booking,
		Guest:   guest,
	})
	return refundAmount, nil
}

// GetBooking returns booking by ID
func (s *BookingService) GetBooking(id string) (*models.Booking, error) {
	return s.bookingRepo.GetByID(id)
}
