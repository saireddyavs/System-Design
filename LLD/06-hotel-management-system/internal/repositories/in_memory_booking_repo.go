package repositories

import (
	"errors"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
	"time"
)

var ErrBookingNotFound = errors.New("booking not found")

// InMemoryBookingRepository implements BookingRepository with thread-safe in-memory storage
type InMemoryBookingRepository struct {
	bookings map[string]*models.Booking
	mu       sync.RWMutex
}

// NewInMemoryBookingRepository creates a new in-memory booking repository
func NewInMemoryBookingRepository() interfaces.BookingRepository {
	return &InMemoryBookingRepository{
		bookings: make(map[string]*models.Booking),
	}
}

func (r *InMemoryBookingRepository) Create(booking *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.bookings[booking.ID]; exists {
		return errors.New("booking already exists")
	}
	r.bookings[booking.ID] = booking
	return nil
}

func (r *InMemoryBookingRepository) GetByID(id string) (*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	booking, exists := r.bookings[id]
	if !exists {
		return nil, ErrBookingNotFound
	}
	return booking, nil
}

func (r *InMemoryBookingRepository) GetByGuestID(guestID string) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Booking
	for _, b := range r.bookings {
		if b.GuestID == guestID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *InMemoryBookingRepository) GetByRoomID(roomID string) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Booking
	for _, b := range r.bookings {
		if b.RoomID == roomID {
			result = append(result, b)
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
	r.bookings[booking.ID] = booking
	return nil
}

func (r *InMemoryBookingRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.bookings[id]; !exists {
		return ErrBookingNotFound
	}
	delete(r.bookings, id)
	return nil
}

func (r *InMemoryBookingRepository) GetBookingsForRoomInRange(roomID string, checkIn, checkOut time.Time) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Booking
	for _, b := range r.bookings {
		if b.RoomID != roomID {
			continue
		}
		if b.Status == models.BookingStatusCancelled {
			continue
		}
		// Overlap: (StartA < EndB) and (EndA > StartB)
		if b.CheckInDate.Before(checkOut) && b.CheckOutDate.After(checkIn) {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *InMemoryBookingRepository) GetActiveBookingsForRoom(roomID string, asOf time.Time) ([]*models.Booking, error) {
	return r.GetBookingsForRoomInRange(roomID, asOf, asOf.Add(24*time.Hour))
}
