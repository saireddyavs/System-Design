package repositories

import (
	"errors"
	"library-management-system/internal/models"
	"library-management-system/internal/interfaces"
	"sort"
	"sync"
)

var ErrReservationNotFound = errors.New("reservation not found")

// InMemoryReservationRepo implements ReservationRepository with thread-safe in-memory storage
type InMemoryReservationRepo struct {
	reservations map[string]*models.Reservation
	mu           sync.RWMutex
}

// NewInMemoryReservationRepo creates a new in-memory reservation repository
func NewInMemoryReservationRepo() interfaces.ReservationRepository {
	return &InMemoryReservationRepo{
		reservations: make(map[string]*models.Reservation),
	}
}

func (r *InMemoryReservationRepo) Create(reservation *models.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reservations[reservation.ID] = reservation
	return nil
}

func (r *InMemoryReservationRepo) GetByID(id string) (*models.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.reservations[id]
	if !ok {
		return nil, ErrReservationNotFound
	}
	return res, nil
}

func (r *InMemoryReservationRepo) Update(reservation *models.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reservations[reservation.ID]; !ok {
		return ErrReservationNotFound
	}
	r.reservations[reservation.ID] = reservation
	return nil
}

func (r *InMemoryReservationRepo) GetPendingByBookID(bookID string) ([]*models.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Reservation
	for _, res := range r.reservations {
		if res.BookID == bookID && res.Status == models.ReservationStatusPending {
			result = append(result, res)
		}
	}
	// Sort by ReservedAt (FIFO queue)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ReservedAt.Before(result[j].ReservedAt)
	})
	return result, nil
}

func (r *InMemoryReservationRepo) GetByMemberID(memberID string) ([]*models.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Reservation
	for _, res := range r.reservations {
		if res.MemberID == memberID {
			result = append(result, res)
		}
	}
	return result, nil
}
