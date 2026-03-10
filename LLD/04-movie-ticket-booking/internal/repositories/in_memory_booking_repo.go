package repositories

import (
	"errors"
	"movie-ticket-booking/internal/models"
	"sync"
)

var ErrBookingNotFound = errors.New("booking not found")

// InMemoryBookingRepository implements BookingRepository
type InMemoryBookingRepository struct {
	bookings map[string]*models.Booking
	mu       sync.RWMutex
}

// NewInMemoryBookingRepository creates a new in-memory booking repository
func NewInMemoryBookingRepository() *InMemoryBookingRepository {
	return &InMemoryBookingRepository{
		bookings: make(map[string]*models.Booking),
	}
}

func (r *InMemoryBookingRepository) Create(booking *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bookings[booking.ID] = booking
	return nil
}

func (r *InMemoryBookingRepository) GetByID(id string) (*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bookings[id]
	if !ok {
		return nil, ErrBookingNotFound
	}
	return b, nil
}

func (r *InMemoryBookingRepository) GetByUserID(userID string) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Booking
	for _, b := range r.bookings {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *InMemoryBookingRepository) GetByShowID(showID string) ([]*models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Booking
	for _, b := range r.bookings {
		if b.ShowID == showID && b.Status == models.BookingStatusConfirmed {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *InMemoryBookingRepository) Update(booking *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bookings[booking.ID]; !ok {
		return ErrBookingNotFound
	}
	r.bookings[booking.ID] = booking
	return nil
}
