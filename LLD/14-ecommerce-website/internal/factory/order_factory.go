package factory

import (
	"time"

	"ecommerce-website/internal/models"
)

// OrderFactory creates orders from cart (Factory pattern)
type OrderFactory struct {
	idGenerator func() string
}

// NewOrderFactory creates a new order factory
func NewOrderFactory(idGenerator func() string) *OrderFactory {
	return &OrderFactory{
		idGenerator: idGenerator,
	}
}

// CreateOrderInput contains data needed to create an order
type CreateOrderInput struct {
	UserID          string
	CartItems       map[string]models.CartItem
	ProductDetails  map[string]*models.Product
	ShippingAddress models.Address
	CouponCode      string
	Discount        float64
	PaymentID       string
}

// Create builds an Order from cart and other inputs
func (f *OrderFactory) Create(input CreateOrderInput) *models.Order {
	now := time.Now()
	items := make([]models.OrderItem, 0, len(input.CartItems))
	totalAmount := 0.0
	totalQty := 0

	for productID, cartItem := range input.CartItems {
		product := input.ProductDetails[productID]
		name := ""
		if product != nil {
			name = product.Name
		}
		items = append(items, models.OrderItem{
			ProductID: productID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
			Name:      name,
		})
		totalAmount += cartItem.Price * float64(cartItem.Quantity)
		totalQty += cartItem.Quantity
	}

	finalAmount := totalAmount - input.Discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &models.Order{
		ID:              f.idGenerator(),
		UserID:          input.UserID,
		Items:           items,
		TotalAmount:     totalAmount,
		Discount:        input.Discount,
		FinalAmount:     finalAmount,
		CouponCode:      input.CouponCode,
		Status:          models.OrderStatusPlaced,
		ShippingAddress: input.ShippingAddress,
		PaymentID:       input.PaymentID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
