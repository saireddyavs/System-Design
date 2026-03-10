package repositories

import (
	"context"
	"strings"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryCouponRepo implements CouponRepository with thread-safe in-memory storage
type InMemoryCouponRepo struct {
	mu      sync.RWMutex
	coupons map[string]*models.Coupon
	byCode  map[string]string
}

// NewInMemoryCouponRepo creates a new in-memory coupon repository
func NewInMemoryCouponRepo() *InMemoryCouponRepo {
	return &InMemoryCouponRepo{
		coupons: make(map[string]*models.Coupon),
		byCode:  make(map[string]string),
	}
}

var _ interfaces.CouponRepository = (*InMemoryCouponRepo)(nil)

func (r *InMemoryCouponRepo) Create(ctx context.Context, coupon *models.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.coupons[coupon.ID]; exists {
		return ErrAlreadyExists
	}
	code := normalizeCode(coupon.Code)
	if _, exists := r.byCode[code]; exists {
		return ErrAlreadyExists
	}
	r.coupons[coupon.ID] = coupon
	r.byCode[code] = coupon.ID
	return nil
}

func (r *InMemoryCouponRepo) GetByID(ctx context.Context, id string) (*models.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.coupons[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *InMemoryCouponRepo) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byCode[normalizeCode(code)]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *r.coupons[id]
	return &cp, nil
}

func (r *InMemoryCouponRepo) Update(ctx context.Context, coupon *models.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.coupons[coupon.ID]; !ok {
		return ErrNotFound
	}
	r.coupons[coupon.ID] = coupon
	return nil
}

func (r *InMemoryCouponRepo) IncrementUsage(ctx context.Context, couponID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.coupons[couponID]
	if !ok {
		return ErrNotFound
	}
	c.UsedCount++
	return nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
