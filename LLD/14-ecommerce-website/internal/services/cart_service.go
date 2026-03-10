package services

import (
	"context"
	"time"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
)

// CartService handles shopping cart operations
type CartService struct {
	cartRepo    interfaces.CartRepository
	productRepo interfaces.ProductRepository
	idGen       func() string
}

// NewCartService creates a new cart service
func NewCartService(cartRepo interfaces.CartRepository, productRepo interfaces.ProductRepository, idGen func() string) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		idGen:       idGen,
	}
}

// GetOrCreateCart returns the user's cart or creates one if not exists
func (s *CartService) GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err == nil {
		return cart, nil
	}
	cart = &models.Cart{
		ID:        s.idGen(),
		UserID:    userID,
		Items:     make(map[string]models.CartItem),
		UpdatedAt: time.Now(),
	}
	if err := s.cartRepo.Create(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// AddItem adds or updates an item in the cart
func (s *CartService) AddItem(ctx context.Context, userID, productID string, quantity int) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < quantity {
		return repositories.ErrInsufficientStock
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	item, exists := cart.Items[productID]
	if exists {
		newQty := item.Quantity + quantity
		if product.Stock < newQty {
			return repositories.ErrInsufficientStock
		}
		item.Quantity = newQty
	} else {
		item = models.CartItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     product.Price,
		}
	}
	cart.Items[productID] = item
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(ctx, cart)
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(ctx context.Context, userID, productID string) error {
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}
	delete(cart.Items, productID)
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(ctx, cart)
}

// UpdateQuantity updates the quantity of an item
func (s *CartService) UpdateQuantity(ctx context.Context, userID, productID string, quantity int) error {
	if quantity <= 0 {
		return s.RemoveItem(ctx, userID, productID)
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < quantity {
		return repositories.ErrInsufficientStock
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	item, exists := cart.Items[productID]
	if !exists {
		cart.Items[productID] = models.CartItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     product.Price,
		}
	} else {
		item.Quantity = quantity
		item.Price = product.Price
		cart.Items[productID] = item
	}
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(ctx, cart)
}

// GetCart returns the user's cart
func (s *CartService) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	return s.GetOrCreateCart(ctx, userID)
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}
	cart.Items = make(map[string]models.CartItem)
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(ctx, cart)
}
