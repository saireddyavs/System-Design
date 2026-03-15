package models

import "time"

// OrderStatus represents the lifecycle of an order.
type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusPaid    OrderStatus = "paid"
)

// OrderItem represents a line item in an order.
type OrderItem struct {
	BookID   string  `json:"book_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// Order represents a placed order.
type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	Items         []OrderItem `json:"items"`
	TotalAmount   float64     `json:"total_amount"`
	Status        OrderStatus `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	CreatedAt     time.Time  `json:"created_at"`
}
