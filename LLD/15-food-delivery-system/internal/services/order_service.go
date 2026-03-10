package services

import (
	"context"
	"errors"
	"fmt"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
	"sync"
	"time"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrRestaurantClosed   = errors.New("restaurant is closed")
	ErrItemUnavailable    = errors.New("menu item not available")
	ErrMinOrderNotMet     = errors.New("order does not meet minimum order amount")
	ErrInvalidTransition  = errors.New("invalid order status transition")
	ErrOrderCannotCancel  = errors.New("order cannot be cancelled at this stage")
)

const MaxDeliveryRadiusKm = 5.0

// OrderService handles order business logic (Factory for order creation, State for lifecycle)
type OrderService struct {
	orderRepo       interfaces.OrderRepository
	restaurantRepo  interfaces.RestaurantRepository
	customerRepo    interfaces.CustomerRepository
	deliveryService *DeliveryService
	paymentService  *PaymentService
	pricingStrategy interfaces.PricingStrategy
	observerManager *OrderObserverManager
	orderIDCounter  int
	mu              sync.Mutex
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo interfaces.OrderRepository,
	restaurantRepo interfaces.RestaurantRepository,
	customerRepo interfaces.CustomerRepository,
	deliveryService *DeliveryService,
	paymentService *PaymentService,
	pricingStrategy interfaces.PricingStrategy,
	observerManager *OrderObserverManager,
) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		restaurantRepo:  restaurantRepo,
		customerRepo:    customerRepo,
		deliveryService: deliveryService,
		paymentService:  paymentService,
		pricingStrategy: pricingStrategy,
		observerManager: observerManager,
		orderIDCounter:  0,
	}
}

// PlaceOrder creates and places a new order (Factory Pattern)
func (s *OrderService) PlaceOrder(ctx context.Context, customerID, restaurantID string, items []models.OrderItem, deliveryAddr models.Location) (*models.Order, error) {
	// Validate customer
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}
	_ = customer // used for validation

	// Validate restaurant
	restaurant, err := s.restaurantRepo.GetByID(restaurantID)
	if err != nil {
		return nil, err
	}
	if !restaurant.IsOpen {
		return nil, ErrRestaurantClosed
	}

	// Validate items and calculate subtotal
	menu := restaurant.GetMenu()
	menuMap := make(map[string]models.MenuItem)
	for _, m := range menu {
		menuMap[m.ID] = m
	}

	var subTotal float64
	for i := range items {
		item := &items[i]
		menuItem, ok := menuMap[item.MenuItemID]
		if !ok {
			return nil, fmt.Errorf("menu item %s not found", item.MenuItemID)
		}
		if !menuItem.IsAvailable {
			return nil, ErrItemUnavailable
		}
		item.Name = menuItem.Name
		item.Price = menuItem.Price
		subTotal += item.Price * float64(item.Quantity)
	}

	if subTotal < restaurant.MinOrder {
		return nil, ErrMinOrderNotMet
	}

	// Generate order ID
	s.mu.Lock()
	s.orderIDCounter++
	orderID := fmt.Sprintf("ORD-%d", s.orderIDCounter)
	s.mu.Unlock()

	// Create order (Factory)
	order := models.NewOrder(orderID, customerID, restaurantID, items, deliveryAddr)

	// Calculate pricing
	deliveryFee := s.pricingStrategy.CalculateDeliveryFee(restaurant.Location, deliveryAddr)
	surgeMultiplier := s.pricingStrategy.CalculateSurgeFee(time.Now())
	surgeAmount := subTotal * surgeMultiplier
	total := s.pricingStrategy.CalculateTotal(subTotal, deliveryFee, surgeAmount)

	order.SetAmounts(subTotal, deliveryFee, surgeAmount, total)

	// Assign delivery agent
	agent, err := s.deliveryService.AssignAgent(restaurant.Location, deliveryAddr)
	if err != nil {
		return nil, err
	}
	order.AssignAgent(agent.ID)
	agent.SetStatus(models.AgentStatusOnDelivery)

	// Persist order
	if err := s.orderRepo.Create(order); err != nil {
		agent.SetStatus(models.AgentStatusAvailable)
		return nil, err
	}

	// Process payment
	payment := models.NewPayment(fmt.Sprintf("PAY-%s", orderID), order.ID, order.Total, models.PaymentMethodUPI)
	if err := s.paymentService.ProcessPayment(ctx, payment); err != nil {
		order.SetStatus(models.OrderStatusCancelled)
		s.deliveryService.MarkAgentAvailable(agent.ID)
		s.orderRepo.Update(order)
		return nil, err
	}

	// Confirm order
	oldStatus := order.Status
	order.SetStatus(models.OrderStatusConfirmed)
	s.orderRepo.Update(order)
	s.observerManager.NotifyStatusChanged(order, oldStatus, models.OrderStatusConfirmed)

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	return s.orderRepo.GetByID(id)
}

// UpdateOrderStatus transitions order to new status (State Pattern)
func (s *OrderService) UpdateOrderStatus(id string, newStatus models.OrderStatus) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	oldStatus := order.Status
	if !order.CanTransition(newStatus) {
		return ErrInvalidTransition
	}

	order.SetStatus(newStatus)

	// Update agent status when order is delivered
	if newStatus == models.OrderStatusDelivered && order.AgentID != "" {
		s.deliveryService.MarkAgentAvailable(order.AgentID)
	}

	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	s.observerManager.NotifyStatusChanged(order, oldStatus, newStatus)
	return nil
}

// CancelOrder cancels an order if allowed
func (s *OrderService) CancelOrder(id string) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !order.CanCancel() {
		return ErrOrderCannotCancel
	}

	oldStatus := order.Status
	order.SetStatus(models.OrderStatusCancelled)

	if order.AgentID != "" {
		s.deliveryService.MarkAgentAvailable(order.AgentID)
	}

	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	s.observerManager.NotifyStatusChanged(order, oldStatus, models.OrderStatusCancelled)
	return nil
}

// GetOrderTracking returns current order status for real-time tracking
func (s *OrderService) GetOrderTracking(id string) (*models.Order, error) {
	return s.orderRepo.GetByID(id)
}
