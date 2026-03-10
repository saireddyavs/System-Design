package models

import (
	"sync"
	"time"
)

// OrderStatus represents the lifecycle state of an order
type OrderStatus string

const (
	OrderStatusPlaced    OrderStatus = "placed"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusPickedUp  OrderStatus = "picked_up"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item in an order
type OrderItem struct {
	MenuItemID string  `json:"menu_item_id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

// Order represents a food delivery order
type Order struct {
	ID           string      `json:"id"`
	CustomerID   string      `json:"customer_id"`
	RestaurantID string      `json:"restaurant_id"`
	AgentID      string      `json:"agent_id"`
	Items        []OrderItem `json:"items"`
	DeliveryAddr Location    `json:"delivery_addr"`
	SubTotal     float64     `json:"sub_total"`
	DeliveryFee  float64     `json:"delivery_fee"`
	SurgeFee     float64     `json:"surge_fee"`
	Total        float64     `json:"total"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	mu           sync.RWMutex
}

// NewOrder creates a new order
func NewOrder(id, customerID, restaurantID string, items []OrderItem, deliveryAddr Location) *Order {
	now := time.Now()
	return &Order{
		ID:           id,
		CustomerID:   customerID,
		RestaurantID: restaurantID,
		Items:        items,
		DeliveryAddr: deliveryAddr,
		Status:       OrderStatusPlaced,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// CanTransition checks if a status transition is valid
func (o *Order) CanTransition(to OrderStatus) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return canTransition(o.Status, to)
}

// CanCancel returns true if order can be cancelled (before preparation starts)
func (o *Order) CanCancel() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.Status == OrderStatusPlaced || o.Status == OrderStatusConfirmed
}

// SetStatus updates the order status
func (o *Order) SetStatus(status OrderStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if canTransition(o.Status, status) {
		o.Status = status
		o.UpdatedAt = time.Now()
	}
}

// AssignAgent assigns a delivery agent to the order
func (o *Order) AssignAgent(agentID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.AgentID = agentID
}

// SetAmounts sets the order amounts
func (o *Order) SetAmounts(subTotal, deliveryFee, surgeFee, total float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.SubTotal = subTotal
	o.DeliveryFee = deliveryFee
	o.SurgeFee = surgeFee
	o.Total = total
}

// canTransition defines the valid order state machine transitions
func canTransition(from, to OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPlaced:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed: {OrderStatusPreparing, OrderStatusCancelled},
		OrderStatusPreparing: {OrderStatusPickedUp},
		OrderStatusPickedUp:  {OrderStatusDelivered},
		OrderStatusDelivered: {},
		OrderStatusCancelled: {},
	}
	valid, ok := transitions[from]
	if !ok {
		return false
	}
	for _, v := range valid {
		if v == to {
			return true
		}
	}
	return false
}
