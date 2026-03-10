package repositories

import (
	"errors"
	"ride-sharing-service/internal/models"
	"sync"
)

var ErrPaymentNotFound = errors.New("payment not found")

// InMemoryPaymentRepository implements PaymentRepository with thread-safe in-memory storage
type InMemoryPaymentRepository struct {
	payments map[string]*models.Payment
	mu       sync.RWMutex
}

// NewInMemoryPaymentRepository creates a new in-memory payment repository
func NewInMemoryPaymentRepository() *InMemoryPaymentRepository {
	return &InMemoryPaymentRepository{
		payments: make(map[string]*models.Payment),
	}
}

// Create adds a new payment
func (r *InMemoryPaymentRepository) Create(payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.payments[payment.ID] = payment
	return nil
}

// GetByID retrieves a payment by ID
func (r *InMemoryPaymentRepository) GetByID(id string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	payment, ok := r.payments[id]
	if !ok {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

// GetByRideID retrieves payment for a ride
func (r *InMemoryPaymentRepository) GetByRideID(rideID string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.payments {
		if p.RideID == rideID {
			return p, nil
		}
	}
	return nil, ErrPaymentNotFound
}

// Update updates an existing payment
func (r *InMemoryPaymentRepository) Update(payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.payments[payment.ID]; !ok {
		return ErrPaymentNotFound
	}
	r.payments[payment.ID] = payment
	return nil
}
