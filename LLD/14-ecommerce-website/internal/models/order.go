package models

import "time"

// OrderStatus represents the lifecycle state of an order
type OrderStatus string

const (
	OrderStatusPlaced    OrderStatus = "Placed"
	OrderStatusConfirmed OrderStatus = "Confirmed"
	OrderStatusCancelled OrderStatus = "Cancelled"
)

// OrderItem represents a line item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Name      string  `json:"name"`
}

// Order represents a customer order
type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Items           []OrderItem `json:"items"`
	TotalAmount     float64     `json:"total_amount"`
	Discount        float64     `json:"discount"`
	FinalAmount     float64     `json:"final_amount"`
	Status          OrderStatus `json:"status"`
	ShippingAddress Address     `json:"shipping_address"`
	PaymentID       string     `json:"payment_id"`
	CouponCode      string     `json:"coupon_code,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
