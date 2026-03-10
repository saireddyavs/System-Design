package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

var (
	ErrCartNotFound   = errors.New("cart not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// CartService manages shopping cart operations (SRP).
// DIP: Depends on CartRepository and BookRepository interfaces.
type CartService struct {
	cartRepo interfaces.CartRepository
	bookRepo interfaces.BookRepository
}

func NewCartService(cartRepo interfaces.CartRepository, bookRepo interfaces.BookRepository) *CartService {
	return &CartService{
		cartRepo: cartRepo,
		bookRepo: bookRepo,
	}
}

func (s *CartService) CreateCart(userID string) (*models.Cart, error) {
	cart := &models.Cart{
		ID:        generateCartID(),
		UserID:    userID,
		Items:     make(map[string]int),
		UpdatedAt: time.Now(),
	}
	if err := s.cartRepo.Create(cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCart(userID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil || cart == nil {
		return nil, ErrCartNotFound
	}
	return cart, nil
}

func (s *CartService) AddToCart(userID, bookID string, quantity int) error {
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil || book == nil {
		return errors.New("book not found")
	}
	if book.Stock < quantity {
		return ErrInsufficientStock
	}

	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil || cart == nil {
		return ErrCartNotFound
	}

	if cart.Items == nil {
		cart.Items = make(map[string]int)
	}
	cart.Items[bookID] += quantity
	cart.UpdatedAt = time.Now()

	return s.cartRepo.Update(cart)
}

func (s *CartService) RemoveFromCart(userID, bookID string) error {
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil || cart == nil {
		return ErrCartNotFound
	}
	delete(cart.Items, bookID)
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(cart)
}

func (s *CartService) UpdateQuantity(userID, bookID string, quantity int) error {
	if quantity <= 0 {
		return s.RemoveFromCart(userID, bookID)
	}

	book, err := s.bookRepo.GetByID(bookID)
	if err != nil || book == nil {
		return errors.New("book not found")
	}
	if book.Stock < quantity {
		return ErrInsufficientStock
	}

	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil || cart == nil {
		return ErrCartNotFound
	}
	cart.Items[bookID] = quantity
	cart.UpdatedAt = time.Now()
	return s.cartRepo.Update(cart)
}

func generateCartID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
