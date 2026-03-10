package repositories

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/models"
)

type InMemoryCouponRepository struct {
	mu      sync.RWMutex
	byID    map[string]*models.Coupon
	byCode  map[string]*models.Coupon
}

func NewInMemoryCouponRepository() *InMemoryCouponRepository {
	return &InMemoryCouponRepository{
		byID:   make(map[string]*models.Coupon),
		byCode: make(map[string]*models.Coupon),
	}
}

func (r *InMemoryCouponRepository) GetByCode(code string) (*models.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byCode[code]
	if !ok {
		return nil, fmt.Errorf("coupon not found: %s", code)
	}
	return copyCoupon(c), nil
}

func (r *InMemoryCouponRepository) GetByID(id string) (*models.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("coupon not found: %s", id)
	}
	return copyCoupon(c), nil
}

func (r *InMemoryCouponRepository) Create(coupon *models.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[coupon.ID]; exists {
		return fmt.Errorf("coupon already exists: %s", coupon.ID)
	}
	cpy := copyCoupon(coupon)
	r.byID[coupon.ID] = cpy
	r.byCode[coupon.Code] = cpy
	return nil
}

func (r *InMemoryCouponRepository) Update(coupon *models.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.byID[coupon.ID]
	if !ok {
		return fmt.Errorf("coupon not found: %s", coupon.ID)
	}
	// If code changed, remove old mapping
	if old.Code != coupon.Code {
		delete(r.byCode, old.Code)
	}
	cpy := copyCoupon(coupon)
	r.byID[coupon.ID] = cpy
	r.byCode[coupon.Code] = cpy
	return nil
}

func (r *InMemoryCouponRepository) IncrementUsage(couponID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[couponID]
	if !ok {
		return fmt.Errorf("coupon not found: %s", couponID)
	}
	c.CurrentUsage++
	return nil
}

func copyCoupon(c *models.Coupon) *models.Coupon {
	cpy := *c
	return &cpy
}
