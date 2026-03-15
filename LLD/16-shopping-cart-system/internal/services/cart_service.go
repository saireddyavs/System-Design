package services

import (
	"fmt"
	"sync"
	"time"

	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
)

type CartService struct {
	cartRepo    interfaces.CartRepository
	productRepo interfaces.ProductRepository
	observers   []interfaces.CartEventObserver
	mu          sync.RWMutex
}

func NewCartService(cartRepo interfaces.CartRepository, productRepo interfaces.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		observers:   make([]interfaces.CartEventObserver, 0),
	}
}

func (s *CartService) RegisterObserver(observer interfaces.CartEventObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *CartService) notify(event interfaces.CartEvent) {
	s.mu.RLock()
	obs := make([]interfaces.CartEventObserver, len(s.observers))
	copy(obs, s.observers)
	s.mu.RUnlock()
	for _, o := range obs {
		o.OnCartEvent(event)
	}
}

func (s *CartService) getOrCreateCart(userID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(userID)
	if err == nil {
		return cart, nil
	}
	// Create new cart
	cart = &models.Cart{
		ID:        fmt.Sprintf("cart_%d", time.Now().UnixNano()),
		UserID:    userID,
		Items:     []models.CartItem{},
		Status:    models.CartStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.cartRepo.Create(cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) AddItem(userID, productID string, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return err
	}
	if !product.HasStock(quantity) {
		return fmt.Errorf("insufficient stock: have %d, requested %d", product.Stock, quantity)
	}
	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return err
	}
	// Check existing quantity + new quantity
	existingQty := 0
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			existingQty = cart.Items[i].Quantity
			break
		}
	}
	totalQty := existingQty + quantity
	if !product.HasStock(totalQty) {
		return fmt.Errorf("insufficient stock: have %d, total requested %d", product.Stock, totalQty)
	}
	// Add or update item
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].RecalculateSubtotal()
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, models.NewCartItem(productID, product.Name, product.Price, quantity))
	}
	cart.UpdatedAt = time.Now()
	if err := s.cartRepo.Update(cart); err != nil {
		return err
	}
	s.notify(interfaces.CartEvent{Type: interfaces.CartEventItemAdded, Cart: cart, UserID: userID})
	return nil
}

func (s *CartService) UpdateQuantity(userID, productID string, quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return err
	}
	if quantity == 0 {
		return s.RemoveItem(userID, productID)
	}
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return err
	}
	if !product.HasStock(quantity) {
		return fmt.Errorf("insufficient stock: have %d, requested %d", product.Stock, quantity)
	}
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity = quantity
			cart.Items[i].RecalculateSubtotal()
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("product not in cart: %s", productID)
	}
	cart.UpdatedAt = time.Now()
	if err := s.cartRepo.Update(cart); err != nil {
		return err
	}
	s.notify(interfaces.CartEvent{Type: interfaces.CartEventItemUpdated, Cart: cart, UserID: userID})
	return nil
}

func (s *CartService) RemoveItem(userID, productID string) error {
	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return err
	}
	newItems := make([]models.CartItem, 0, len(cart.Items))
	found := false
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		} else {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("product not in cart: %s", productID)
	}
	cart.Items = newItems
	cart.UpdatedAt = time.Now()
	if err := s.cartRepo.Update(cart); err != nil {
		return err
	}
	s.notify(interfaces.CartEvent{Type: interfaces.CartEventItemRemoved, Cart: cart, UserID: userID})
	return nil
}

func (s *CartService) GetCart(userID string) (*models.Cart, error) {
	return s.cartRepo.GetByUserID(userID)
}
