package services

import (
	"fmt"
	"time"

	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
)

// OrderFactory creates Order from Cart (Factory pattern)
type OrderFactory struct{}

func (f *OrderFactory) CreateFromCart(cart *models.Cart, user *models.User, subtotal, discount, tax, total float64, paymentID, couponCode string) *models.Order {
	items := make([]models.OrderItem, len(cart.Items))
	for i, ci := range cart.Items {
		items[i] = models.OrderItem{
			ProductID:   ci.ProductID,
			ProductName: ci.ProductName,
			UnitPrice:   ci.UnitPrice,
			Quantity:    ci.Quantity,
			Subtotal:    ci.Subtotal,
		}
	}
	return &models.Order{
		ID:              fmt.Sprintf("ord_%d", time.Now().UnixNano()),
		UserID:          cart.UserID,
		Items:           items,
		Subtotal:        subtotal,
		Discount:        discount,
		Tax:             tax,
		Total:           total,
		Status:          models.OrderStatusConfirmed,
		PaymentID:       paymentID,
		ShippingAddress: user.Address,
		CouponCode:      couponCode,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

type OrderService struct {
	repo interfaces.OrderRepository
}

func NewOrderService(repo interfaces.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetByUserID(userID string) ([]*models.Order, error) {
	return s.repo.GetByUserID(userID)
}
