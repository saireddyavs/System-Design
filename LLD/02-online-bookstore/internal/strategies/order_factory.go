package strategies

import (
	"errors"
	"crypto/rand"
	"encoding/hex"
	"time"

	"online-bookstore/internal/models"
)

var (
	ErrEmptyCart       = errors.New("cart is empty")
	ErrBookPriceNotFound = errors.New("book price not found")
)

// OrderFactory creates orders from cart items (Factory pattern).
// Encapsulates order creation logic and ID generation.
type OrderFactory struct{}

func NewOrderFactory() *OrderFactory {
	return &OrderFactory{}
}

// CreateOrder builds an Order from cart items and user.
// Factory: Centralizes order object creation with validation.
func (f *OrderFactory) CreateOrder(userID string, items map[string]int, bookPrices map[string]float64, paymentMethod string) (*models.Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyCart
	}

	orderItems := make([]models.OrderItem, 0, len(items))
	var totalAmount float64

	for bookID, qty := range items {
		price, ok := bookPrices[bookID]
		if !ok {
			return nil, ErrBookPriceNotFound
		}
		orderItems = append(orderItems, models.OrderItem{
			BookID:   bookID,
			Quantity: qty,
			Price:    price,
		})
		totalAmount += price * float64(qty)
	}

	return &models.Order{
		ID:            generateID(),
		UserID:        userID,
		Items:         orderItems,
		TotalAmount:   totalAmount,
		Status:        models.OrderStatusPending,
		PaymentMethod: paymentMethod,
		CreatedAt:     time.Now(),
	}, nil
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
