package repositories

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
	"sync"
)

var (
	ErrBookingNotFound = errors.New("booking not found")
)

// InMemoryBookingRepository implements BookingRepository with in-memory storage (thread-safe)
type InMemoryBookingRepository struct {
	bookings       map[string]*models.Booking
	bookingRefIdx  map[string]string // bookingRef -> bookingID
	passengerIdx   map[string][]string
	flightIdx      map[string][]string
	mu             sync.RWMutex
}

// NewInMemoryBookingRepository creates a new in-memory booking repository
func NewInMemoryBookingRepository() interfaces.BookingRepository {
	return &InMemoryBookingRepository{
		bookings:      make(map[string]*models.Booking),
		bookingRefIdx: make(map[string]string),
		passengerIdx:  make(map[string][]string),
		flightIdx:     make(map[string][]string),
	}
}

func (r *InMemoryBookingRepository) Create(booking *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookings[booking.ID]; exists {
		return errors.New("booking already exists")
	}
	if _, exists := r.bookingRefIdx[booking.BookingRef]; exists {
		return errors.New("booking reference already exists")
	}

	bookingCopy := copyBooking(booking)
	r.bookings[booking.ID] = bookingCopy
	r.bookingRefIdx[booking.BookingRef] = booking.ID
	r.passengerIdx[booking.PassengerID] = append(r.passengerIdx[booking.PassengerID], booking.ID)
	r.flightIdx[booking.FlightID] = append(r.flightIdx[booking.FlightID], booking.ID)
	return nil
}

func (r *InMemoryBookingRepository) GetByID(id string) (*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	booking, exists := r.bookings[id]
	if !exists {
		return nil, ErrBookingNotFound
	}
	return copyBooking(booking), nil
}

func (r *InMemoryBookingRepository) GetByBookingRef(bookingRef string) (*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bookingID, exists := r.bookingRefIdx[bookingRef]
	if !exists {
		return nil, ErrBookingNotFound
	}
	booking := r.bookings[bookingID]
	return copyBooking(booking), nil
}

func (r *InMemoryBookingRepository) GetByFlightID(flightID string) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bookingIDs, exists := r.flightIdx[flightID]
	if !exists {
		return []*models.Booking{}, nil
	}

	result := make([]*models.Booking, 0, len(bookingIDs))
	for _, id := range bookingIDs {
		if b, ok := r.bookings[id]; ok && b.Status == models.BookingStatusConfirmed {
			result = append(result, copyBooking(b))
		}
	}
	return result, nil
}

func (r *InMemoryBookingRepository) Update(booking *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookings[booking.ID]; !exists {
		return ErrBookingNotFound
	}
	r.bookings[booking.ID] = copyBooking(booking)
	return nil
}

func copyBooking(b *models.Booking) *models.Booking {
	seatIDs := make([]string, len(b.SeatIDs))
	copy(seatIDs, b.SeatIDs)
	return &models.Booking{
		ID:          b.ID,
		PassengerID: b.PassengerID,
		FlightID:    b.FlightID,
		SeatIDs:     seatIDs,
		TotalAmount: b.TotalAmount,
		Status:      b.Status,
		BookingRef:  b.BookingRef,
		CreatedAt:   b.CreatedAt,
	}
}
