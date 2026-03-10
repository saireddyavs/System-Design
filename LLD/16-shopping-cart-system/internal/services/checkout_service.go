package services

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/strategies"
)

// CheckoutSummaryBuilder builds checkout result (Builder pattern)
type CheckoutSummaryBuilder struct {
	summary *interfaces.CheckoutSummary
}

func NewCheckoutSummaryBuilder() *CheckoutSummaryBuilder {
	return &CheckoutSummaryBuilder{
		summary: &interfaces.CheckoutSummary{},
	}
}

func (b *CheckoutSummaryBuilder) SetOrder(order *models.Order) *CheckoutSummaryBuilder {
	b.summary.Order = order
	b.summary.OrderID = order.ID
	b.summary.Subtotal = order.Subtotal
	b.summary.Discount = order.Discount
	b.summary.Tax = order.Tax
	b.summary.Total = order.Total
	b.summary.PaymentID = order.PaymentID
	return b
}

func (b *CheckoutSummaryBuilder) Build() *interfaces.CheckoutSummary {
	return b.summary
}

type CheckoutService struct {
	cartRepo       interfaces.CartRepository
	orderRepo      interfaces.OrderRepository
	productRepo    interfaces.ProductRepository
	couponRepo     interfaces.CouponRepository
	userRepo       interfaces.UserRepository
	cartService    *CartService
	couponService  *CouponService
	orderService   *OrderService
	taxCalculator  interfaces.TaxCalculator
	paymentRegistry *strategies.PaymentProcessorRegistry
	mu             sync.Mutex
}

func NewCheckoutService(
	cartRepo interfaces.CartRepository,
	orderRepo interfaces.OrderRepository,
	productRepo interfaces.ProductRepository,
	couponRepo interfaces.CouponRepository,
	userRepo interfaces.UserRepository,
	cartService *CartService,
	couponService *CouponService,
	orderService *OrderService,
	taxCalculator interfaces.TaxCalculator,
	paymentRegistry *strategies.PaymentProcessorRegistry,
) *CheckoutService {
	return &CheckoutService{
		cartRepo:        cartRepo,
		orderRepo:       orderRepo,
		productRepo:     productRepo,
		couponRepo:      couponRepo,
		userRepo:        userRepo,
		cartService:     cartService,
		couponService:   couponService,
		orderService:    orderService,
		taxCalculator:   taxCalculator,
		paymentRegistry: paymentRegistry,
	}
}

func (s *CheckoutService) Checkout(userID string, couponCode string, paymentMethod models.PaymentMethod) (*interfaces.CheckoutSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Get and validate cart
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %w", err)
	}
	if cart.IsEmpty() {
		return nil, fmt.Errorf("cart is empty")
	}
	if cart.Status != models.CartStatusActive {
		return nil, fmt.Errorf("cart is not active for checkout")
	}

	// 2. Stock validation
	productIDs := make([]string, len(cart.Items))
	for i, item := range cart.Items {
		productIDs[i] = item.ProductID
	}
	products, err := s.productRepo.GetByIDs(productIDs)
	if err != nil {
		return nil, err
	}
	productMap := make(map[string]*models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}
	for _, item := range cart.Items {
		p, ok := productMap[item.ProductID]
		if !ok {
			return nil, fmt.Errorf("product not found: %s", item.ProductID)
		}
		if !p.HasStock(item.Quantity) {
			return nil, fmt.Errorf("insufficient stock for %s: have %d, need %d", item.ProductName, p.Stock, item.Quantity)
		}
	}

	// 3. Calculate subtotal
	subtotal := cart.Subtotal()

	// 4. Apply coupon/discount
	var discount float64
	var coupon *models.Coupon
	if couponCode != "" {
		discount, coupon, err = s.couponService.ValidateAndGetDiscount(couponCode, subtotal, cart.Items)
		if err != nil {
			return nil, err
		}
	}

	subtotalAfterDiscount := subtotal - discount
	if subtotalAfterDiscount < 0 {
		subtotalAfterDiscount = 0
	}

	// 5. Calculate tax
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	taxCtx := &interfaces.TaxContext{
		SubtotalAfterDiscount: subtotalAfterDiscount,
		State:                  user.Address.State,
		Country:                user.Address.Country,
	}
	tax := s.taxCalculator.Calculate(taxCtx)

	// 6. Total
	total := subtotalAfterDiscount + tax

	// 7. Process payment
	processor, err := s.paymentRegistry.GetProcessor(paymentMethod)
	if err != nil {
		return nil, err
	}
	paymentReq := &interfaces.PaymentRequest{
		Amount:        total,
		Currency:      "USD",
		PaymentMethod: paymentMethod,
		UserID:        userID,
		OrderID:       "", // set after order creation
		Metadata:      nil,
	}
	// Create order first to get ID for payment
	orderFactory := &OrderFactory{}
	order := orderFactory.CreateFromCart(cart, user, subtotal, discount, tax, total, "", couponCode)
	paymentReq.OrderID = order.ID
	result, err := processor.Process(paymentReq)
	if err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("payment failed: %s", result.Message)
	}
	order.PaymentID = result.PaymentID

	// 8. Decrement stock
	for _, item := range cart.Items {
		if err := s.productRepo.DecrementStock(item.ProductID, item.Quantity); err != nil {
			return nil, fmt.Errorf("stock update failed: %w", err)
		}
	}

	// 9. Create order
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("order creation failed: %w", err)
	}

	// 10. Increment coupon usage
	if coupon != nil {
		_ = s.couponRepo.IncrementUsage(coupon.ID)
	}

	// 11. Clear cart - create new empty cart for user
	newCart := &models.Cart{
		ID:        fmt.Sprintf("cart_%d", order.CreatedAt.UnixNano()+1),
		UserID:    userID,
		Items:     []models.CartItem{},
		Status:    models.CartStatusActive,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.CreatedAt,
	}
	_ = s.cartRepo.UpdateStatus(cart.ID, models.CartStatusCheckedOut)
	_ = s.cartRepo.Create(newCart)

	// Notify observers
	s.cartService.notify(interfaces.CartEvent{Type: interfaces.CartEventCheckout, Cart: cart, UserID: userID})

	// Build summary
	summary := NewCheckoutSummaryBuilder().SetOrder(order).Build()
	return summary, nil
}
