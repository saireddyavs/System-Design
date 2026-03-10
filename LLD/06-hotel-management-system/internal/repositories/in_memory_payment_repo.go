package repositories

import (
	"errors"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
)

var ErrPaymentNotFound = errors.New("payment not found")

// InMemoryPaymentRepository implements PaymentRepository
type InMemoryPaymentRepository struct {
	payments map[string]*models.Payment
	mu       sync.RWMutex
}

// NewInMemoryPaymentRepository creates a new in-memory payment repository
func NewInMemoryPaymentRepository() interfaces.PaymentRepository {
	return &InMemoryPaymentRepository{
		payments: make(map[string]*models.Payment),
	}
}

func (r *InMemoryPaymentRepository) Create(payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.payments[payment.ID]; exists {
		return errors.New("payment already exists")
	}
	r.payments[payment.ID] = payment
	return nil
}

func (r *InMemoryPaymentRepository) GetByID(id string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	payment, exists := r.payments[id]
	if !exists {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (r *InMemoryPaymentRepository) GetByBookingID(bookingID string) ([]*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Payment
	for _, p := range r.payments {
		if p.BookingID == bookingID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryPaymentRepository) Update(payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.payments[payment.ID]; !exists {
		return ErrPaymentNotFound
	}
	r.payments[payment.ID] = payment
	return nil
}
