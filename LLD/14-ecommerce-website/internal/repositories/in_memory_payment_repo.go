package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryPaymentRepo implements PaymentRepository with thread-safe in-memory storage
type InMemoryPaymentRepo struct {
	mu       sync.RWMutex
	payments map[string]*models.Payment
	byOrder  map[string]string
}

// NewInMemoryPaymentRepo creates a new in-memory payment repository
func NewInMemoryPaymentRepo() *InMemoryPaymentRepo {
	return &InMemoryPaymentRepo{
		payments: make(map[string]*models.Payment),
		byOrder:  make(map[string]string),
	}
}

var _ interfaces.PaymentRepository = (*InMemoryPaymentRepo)(nil)

func (r *InMemoryPaymentRepo) Create(ctx context.Context, payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.payments[payment.ID]; exists {
		return ErrAlreadyExists
	}
	r.payments[payment.ID] = payment
	r.byOrder[payment.OrderID] = payment.ID
	return nil
}

func (r *InMemoryPaymentRepo) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.payments[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *InMemoryPaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byOrder[orderID]
	if !ok {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *InMemoryPaymentRepo) Update(ctx context.Context, payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.payments[payment.ID]; !ok {
		return ErrNotFound
	}
	r.payments[payment.ID] = payment
	return nil
}
