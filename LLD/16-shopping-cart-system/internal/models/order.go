package models

import "time"

// OrderItem represents a line item in an order (snapshot at checkout)
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitPrice   float64 `json:"unit_price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// Order represents a completed purchase
type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Items           []OrderItem `json:"items"`
	Subtotal        float64     `json:"subtotal"`
	Discount        float64     `json:"discount"`
	Tax             float64     `json:"tax"`
	Total           float64     `json:"total"`
	Status          OrderStatus `json:"status"`
	PaymentID       string      `json:"payment_id"`
	ShippingAddress Address     `json:"shipping_address"`
	CouponCode      string      `json:"coupon_code,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
