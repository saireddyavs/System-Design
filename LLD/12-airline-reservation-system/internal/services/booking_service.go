package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
	"time"
)

var (
	ErrBookingNotFound     = errors.New("booking not found")
	ErrInvalidCancellation = errors.New("cannot cancel booking")
	ErrDoubleBooking       = errors.New("seat already booked")
)

// RefundPolicy: >48h full, 24-48h 75%, <24h 25%
const (
	RefundFullHours   = 48
	RefundPartial75Hours = 24
)

// BookingService handles booking and cancellation operations
type BookingService struct {
	bookingRepo    interfaces.BookingRepository
	flightRepo     interfaces.FlightRepository
	passengerRepo  interfaces.PassengerRepository
	seatService    *SeatService
	pricingStrategy interfaces.PricingStrategy
	paymentProcessor interfaces.PaymentProcessor
	bookingFactory *BookingFactory
	notifier       *BookingNotifier
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo interfaces.BookingRepository,
	flightRepo interfaces.FlightRepository,
	passengerRepo interfaces.PassengerRepository,
	seatService *SeatService,
	pricingStrategy interfaces.PricingStrategy,
	paymentProcessor interfaces.PaymentProcessor,
	factory *BookingFactory,
	notifier *BookingNotifier,
) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		flightRepo:      flightRepo,
		passengerRepo:   passengerRepo,
		seatService:     seatService,
		pricingStrategy: pricingStrategy,
		paymentProcessor: paymentProcessor,
		bookingFactory:  factory,
		notifier:        notifier,
	}
}

// CreateBooking creates a new booking with automatic seat assignment
func (b *BookingService) CreateBooking(passengerID, flightID string, seatCount int, preferredClass models.SeatClass) (*models.Booking, error) {
	// Validate passenger
	_, err := b.passengerRepo.GetByID(passengerID)
	if err != nil {
		return nil, err
	}

	// Get flight
	flight, err := b.flightRepo.GetByID(flightID)
	if err != nil {
		return nil, err
	}
	if flight.Status == models.FlightStatusCancelled {
		return nil, errors.New("flight is cancelled")
	}

	// Auto-assign seats
	seatIDs, err := b.seatService.AutoAssignSeats(flightID, seatCount, preferredClass)
	if err != nil {
		return nil, err
	}

	// Get seat objects for pricing
	seats := getSeatsByIDs(flight.Seats, seatIDs)
	totalAmount := b.pricingStrategy.CalculatePrice(flight.BasePrice, seats)

	// Process payment
	_, err = b.paymentProcessor.ProcessPayment(totalAmount, "USD", "booking")
	if err != nil {
		return nil, err
	}

	// Create booking (Factory)
	booking, err := b.bookingFactory.CreateBooking(passengerID, flightID, seatIDs, totalAmount)
	if err != nil {
		return nil, err
	}

	// Mark seats as booked (atomic - could use transaction in production)
	if err := b.seatService.MarkSeatsBooked(flightID, seatIDs); err != nil {
		return nil, err
	}

	if err := b.bookingRepo.Create(booking); err != nil {
		// Rollback seat assignment
		_ = b.seatService.ReleaseSeats(flightID, seatIDs)
		return nil, err
	}

	// Notify observers
	b.notifier.NotifyBookingCreated(booking)
	return booking, nil
}

// CreateBookingWithSeats creates a booking with manually specified seats
func (b *BookingService) CreateBookingWithSeats(passengerID, flightID string, seatIDs []string) (*models.Booking, error) {
	_, err := b.passengerRepo.GetByID(passengerID)
	if err != nil {
		return nil, err
	}

	flight, err := b.flightRepo.GetByID(flightID)
	if err != nil {
		return nil, err
	}
	if flight.Status == models.FlightStatusCancelled {
		return nil, errors.New("flight is cancelled")
	}

	// Validate seats are available
	if err := b.seatService.ManualAssignSeats(flightID, seatIDs); err != nil {
		return nil, err
	}

	seats := getSeatsByIDs(flight.Seats, seatIDs)
	totalAmount := b.pricingStrategy.CalculatePrice(flight.BasePrice, seats)

	_, err = b.paymentProcessor.ProcessPayment(totalAmount, "USD", "booking")
	if err != nil {
		return nil, err
	}

	booking, err := b.bookingFactory.CreateBooking(passengerID, flightID, seatIDs, totalAmount)
	if err != nil {
		return nil, err
	}

	if err := b.seatService.MarkSeatsBooked(flightID, seatIDs); err != nil {
		return nil, err
	}

	if err := b.bookingRepo.Create(booking); err != nil {
		_ = b.seatService.ReleaseSeats(flightID, seatIDs)
		return nil, err
	}

	b.notifier.NotifyBookingCreated(booking)
	return booking, nil
}

// CancelBooking cancels a booking and processes refund based on policy
func (b *BookingService) CancelBooking(bookingID string) (float64, error) {
	booking, err := b.bookingRepo.GetByID(bookingID)
	if err != nil {
		return 0, err
	}
	if booking.Status == models.BookingStatusCancelled {
		return 0, ErrInvalidCancellation
	}

	flight, err := b.flightRepo.GetByID(booking.FlightID)
	if err != nil {
		return 0, err
	}

	hoursUntilDeparture := time.Until(flight.DepartureTime).Hours()
	refundPercent := getRefundPercent(hoursUntilDeparture)
	refundAmount := booking.TotalAmount * refundPercent / 100

	// Process refund
	if refundAmount > 0 {
		_ = b.paymentProcessor.ProcessRefund("TXN-booking-0", refundAmount) // Simplified for mock
	}

	// Release seats
	_ = b.seatService.ReleaseSeats(booking.FlightID, booking.SeatIDs)

	// Update booking status
	booking.Status = models.BookingStatusCancelled
	_ = b.bookingRepo.Update(booking)

	b.notifier.NotifyBookingCancelled(booking)
	return refundAmount, nil
}

// GetBooking retrieves a booking by ID or reference
func (b *BookingService) GetBooking(bookingIDOrRef string) (*models.Booking, error) {
	// Try as ID first
	booking, err := b.bookingRepo.GetByID(bookingIDOrRef)
	if err == nil {
		return booking, nil
	}
	// Try as booking reference
	return b.bookingRepo.GetByBookingRef(bookingIDOrRef)
}

func getRefundPercent(hoursUntilDeparture float64) float64 {
	switch {
	case hoursUntilDeparture > RefundFullHours:
		return 100
	case hoursUntilDeparture > RefundPartial75Hours:
		return 75
	default:
		return 25
	}
}

func getSeatsByIDs(seats []*models.Seat, ids []string) []*models.Seat {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	var result []*models.Seat
	for _, s := range seats {
		if idSet[s.ID] {
			result = append(result, s)
		}
	}
	return result
}
