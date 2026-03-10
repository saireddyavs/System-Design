package models

import "time"

// CouponType represents the type of discount
type CouponType string

const (
	CouponTypePercentage CouponType = "Percentage"
	CouponTypeFlat       CouponType = "Flat"
	CouponTypeBOGO       CouponType = "BuyOneGetOne"
)

// Coupon represents a discount coupon
type Coupon struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	Type           CouponType `json:"type"`
	Value          float64    `json:"value"`           // Percentage (0-100) or flat amount
	MinOrderAmount float64    `json:"min_order_amount"`
	ExpiresAt      time.Time  `json:"expires_at"`
	UsageLimit     int        `json:"usage_limit"`
	UsedCount      int        `json:"used_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
