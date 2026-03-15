package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
	"online-bookstore/internal/strategies"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrPaymentFailed   = errors.New("payment failed")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
)

// OrderService manages order placement and payment (SRP).
// Uses Factory for order creation, Strategy for payment processing.
type OrderService struct {
	orderRepo    interfaces.OrderRepository
	cartRepo     interfaces.CartRepository
	bookRepo     interfaces.BookRepository
	orderFactory *strategies.OrderFactory
	paymentReg   *strategies.PaymentProcessorRegistry
}

func NewOrderService(
	orderRepo interfaces.OrderRepository,
	cartRepo interfaces.CartRepository,
	bookRepo interfaces.BookRepository,
	orderFactory *strategies.OrderFactory,
	paymentReg *strategies.PaymentProcessorRegistry,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		cartRepo:     cartRepo,
		bookRepo:     bookRepo,
		orderFactory: orderFactory,
		paymentReg:   paymentReg,
	}
}

func (s *OrderService) PlaceOrder(userID, paymentMethod string) (*models.Order, error) {
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		return nil, errors.New("cart is empty or not found")
	}

	// Build book prices map
	bookPrices := make(map[string]float64)
	for bookID := range cart.Items {
		book, err := s.bookRepo.GetByID(bookID)
		if err != nil || book == nil {
			return nil, errors.New("book not found: " + bookID)
		}
		if book.Stock < cart.Items[bookID] {
			return nil, errors.New("insufficient stock for book: " + book.Title)
		}
		bookPrices[bookID] = book.Price
	}

	// Factory: Create order
	order, err := s.orderFactory.CreateOrder(userID, cart.Items, bookPrices, paymentMethod)
	if err != nil {
		return nil, err
	}

	// Strategy: Process payment
	processor, ok := s.paymentReg.GetProcessor(paymentMethod)
	if !ok {
		return nil, ErrInvalidPaymentMethod
	}

	payment := &models.Payment{
		ID:      generatePaymentID(),
		OrderID: order.ID,
		Amount:  order.TotalAmount,
		Method:  paymentMethod,
		Status:  models.PaymentStatusPending,
		CreatedAt: time.Now(),
	}

	payment, err = processor.Process(payment)
	if err != nil || payment.Status != models.PaymentStatusCompleted {
		return nil, ErrPaymentFailed
	}

	// Persist order
	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// Update inventory
	for bookID, qty := range cart.Items {
		_ = s.bookRepo.UpdateStock(bookID, -qty)
	}

	// Clear cart
	cart.Items = make(map[string]int)
	_ = s.cartRepo.Update(cart)

	order.Status = models.OrderStatusPaid
	_ = s.orderRepo.Update(order)

	return order, nil
}

func (s *OrderService) GetOrderHistory(userID string) ([]*models.Order, error) {
	return s.orderRepo.GetByUserID(userID)
}

func generatePaymentID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
