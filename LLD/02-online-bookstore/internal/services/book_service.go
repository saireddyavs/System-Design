package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

var ErrBookExists = errors.New("book with ISBN already exists")

// BookService manages book catalog operations (SRP: single responsibility).
// DIP: Depends on BookRepository interface, not concrete implementation.
type BookService struct {
	bookRepo interfaces.BookRepository
}

func NewBookService(bookRepo interfaces.BookRepository) *BookService {
	return &BookService{bookRepo: bookRepo}
}

func (s *BookService) CreateBook(title, author, isbn string, price float64, genre string, stock int) (*models.Book, error) {
	existing, _ := s.bookRepo.GetByISBN(isbn)
	if existing != nil {
		return nil, ErrBookExists
	}

	book := &models.Book{
		ID:        generateID(),
		Title:     title,
		Author:    author,
		ISBN:      isbn,
		Price:     price,
		Genre:     genre,
		Stock:     stock,
		CreatedAt: time.Now(),
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookService) GetBook(id string) (*models.Book, error) {
	return s.bookRepo.GetByID(id)
}

func (s *BookService) ListBooks() ([]*models.Book, error) {
	return s.bookRepo.GetAll()
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
