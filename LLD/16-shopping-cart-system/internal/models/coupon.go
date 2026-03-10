package models

import "time"

// Coupon represents a discount code with validation rules
type Coupon struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	Type           CouponType `json:"type"`
	Value          float64    `json:"value"`           // percentage (0-100) or flat amount
	MinOrderAmount float64    `json:"min_order_amount"`
	ExpiresAt      time.Time  `json:"expires_at"`
	MaxUsageLimit  int        `json:"max_usage_limit"` // 0 = unlimited
	CurrentUsage   int        `json:"current_usage"`
	CreatedAt      time.Time  `json:"created_at"`
}

// IsExpired checks if coupon has passed expiry date
func (c *Coupon) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsUsageExhausted checks if max usage limit reached
func (c *Coupon) IsUsageExhausted() bool {
	if c.MaxUsageLimit == 0 {
		return false
	}
	return c.CurrentUsage >= c.MaxUsageLimit
}

// MeetsMinOrder checks if order amount qualifies for coupon
func (c *Coupon) MeetsMinOrder(amount float64) bool {
	return amount >= c.MinOrderAmount
}
