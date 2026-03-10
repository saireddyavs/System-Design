package services

import (
	"context"
	"errors"

	"ecommerce-website/internal/factory"
	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/observer"
	"ecommerce-website/internal/repositories"
)

// OrderService handles order placement and tracking
type OrderService struct {
	orderRepo       interfaces.OrderRepository
	cartRepo        interfaces.CartRepository
	productRepo     interfaces.ProductRepository
	paymentRepo     interfaces.PaymentRepository
	couponRepo      interfaces.CouponRepository
	orderFactory    *factory.OrderFactory
	paymentService  *PaymentService
	couponService   *CouponService
	orderObserver   *observer.OrderStatusObserver
	idGen           func() string
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo interfaces.OrderRepository,
	cartRepo interfaces.CartRepository,
	productRepo interfaces.ProductRepository,
	paymentRepo interfaces.PaymentRepository,
	couponRepo interfaces.CouponRepository,
	orderFactory *factory.OrderFactory,
	paymentService *PaymentService,
	couponService *CouponService,
	orderObserver *observer.OrderStatusObserver,
	idGen func() string,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		cartRepo:       cartRepo,
		productRepo:    productRepo,
		paymentRepo:    paymentRepo,
		couponRepo:     couponRepo,
		orderFactory:   orderFactory,
		paymentService: paymentService,
		couponService:  couponService,
		orderObserver:  orderObserver,
		idGen:          idGen,
	}
}

// PlaceOrderInput contains data for placing an order
type PlaceOrderInput struct {
	UserID          string
	ShippingAddress models.Address
	CouponCode      string
	PaymentMethod   models.PaymentMethod
}

// PlaceOrder creates an order from the user's cart
func (s *OrderService) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*models.Order, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, input.UserID)
	if err != nil || len(cart.Items) == 0 {
		if err == repositories.ErrNotFound {
			return nil, ErrEmptyCart
		}
		return nil, err
	}

	// 1. Validate stock for all items
	productDetails := make(map[string]*models.Product)
	totalAmount := 0.0
	totalQty := 0
	for productID, item := range cart.Items {
		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, err
		}
		if product.Stock < item.Quantity {
			return nil, repositories.ErrInsufficientStock
		}
		productDetails[productID] = product
		totalAmount += item.Price * float64(item.Quantity)
		totalQty += item.Quantity
	}

	// 2. Apply coupon if provided
	discount := 0.0
	var coupon *models.Coupon
	if input.CouponCode != "" {
		var err error
		discount, coupon, err = s.couponService.ApplyCoupon(ctx, input.CouponCode, totalAmount, totalQty)
		if err != nil {
			return nil, err
		}
	}

	// 3. Create payment record
	payment := &models.Payment{
		ID:      s.idGen(),
		OrderID: "", // Set after order creation
		Amount:  totalAmount - discount,
		Method:  input.PaymentMethod,
		Status:  models.PaymentStatusPending,
	}
	// We need order ID first - create order with temp payment ID
	paymentID := payment.ID

	// 4. Create order via factory
	order := s.orderFactory.Create(factory.CreateOrderInput{
		UserID:          input.UserID,
		CartItems:       cart.Items,
		ProductDetails:  productDetails,
		ShippingAddress: input.ShippingAddress,
		CouponCode:      input.CouponCode,
		Discount:        discount,
		PaymentID:       paymentID,
	})

	// 5. Decrement stock (must succeed before payment)
	for productID, item := range cart.Items {
		if err := s.productRepo.DecrementStock(ctx, productID, item.Quantity); err != nil {
			return nil, err
		}
	}

	// 6. Process payment
	payment.OrderID = order.ID
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		// Rollback stock
		for productID, item := range cart.Items {
			_ = s.productRepo.IncrementStock(ctx, productID, item.Quantity)
		}
		return nil, err
	}

	if err := s.paymentService.ProcessPayment(ctx, payment); err != nil {
		// Rollback stock
		for productID, item := range cart.Items {
			_ = s.productRepo.IncrementStock(ctx, productID, item.Quantity)
		}
		return nil, err
	}

	// 7. Update coupon usage
	if coupon != nil {
		_ = s.couponService.IncrementUsage(ctx, coupon.ID)
	}

	// 8. Persist order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 9. Update order status to Confirmed
	order.Status = models.OrderStatusConfirmed
	_ = s.orderRepo.UpdateStatus(ctx, order.ID, models.OrderStatusConfirmed)

	// 10. Clear cart
	cart.Items = make(map[string]models.CartItem)
	_ = s.cartRepo.Update(ctx, cart)

	// 11. Notify observers
	s.orderObserver.NotifyOrderStatus(ctx, order.ID, order.UserID, models.OrderStatusConfirmed)

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

// GetOrderHistory returns orders for a user
func (s *OrderService) GetOrderHistory(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID, limit, offset)
}

// UpdateOrderStatus updates order status and notifies observers
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return err
	}
	order, _ := s.orderRepo.GetByID(ctx, orderID)
	if order != nil {
		s.orderObserver.NotifyOrderStatus(ctx, orderID, order.UserID, status)
	}
	return nil
}

// CancelOrder cancels an order and restores stock
func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != models.OrderStatusPlaced && order.Status != models.OrderStatusConfirmed {
		return ErrOrderCannotBeCancelled
	}

	// Restore stock
	for _, item := range order.Items {
		_ = s.productRepo.IncrementStock(ctx, item.ProductID, item.Quantity)
	}

	order.Status = models.OrderStatusCancelled
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return err
	}
	s.orderObserver.NotifyOrderStatus(ctx, orderID, order.UserID, models.OrderStatusCancelled)
	return nil
}

var (
	ErrEmptyCart             = errors.New("cart is empty")
	ErrOrderCannotBeCancelled = errors.New("order cannot be cancelled in current status")
)
