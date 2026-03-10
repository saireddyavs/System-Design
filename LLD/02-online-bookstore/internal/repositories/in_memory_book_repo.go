package repositories

import (
	"errors"
	"sync"

	"online-bookstore/internal/models"
)

var ErrBookNotFound = errors.New("book not found")

// InMemoryBookRepository implements BookRepository with thread-safe in-memory storage.
// Uses RWMutex: multiple readers OR single writer for concurrent access.
type InMemoryBookRepository struct {
	books map[string]*models.Book
	mu    sync.RWMutex
}

// NewInMemoryBookRepository creates a new in-memory book repository.
func NewInMemoryBookRepository() *InMemoryBookRepository {
	return &InMemoryBookRepository{
		books: make(map[string]*models.Book),
	}
}

func (r *InMemoryBookRepository) Create(book *models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[book.ID]; exists {
		return errors.New("book already exists")
	}
	r.books[book.ID] = book
	return nil
}

func (r *InMemoryBookRepository) GetByID(id string) (*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, exists := r.books[id]
	if !exists {
		return nil, ErrBookNotFound
	}
	return book, nil
}

func (r *InMemoryBookRepository) GetByISBN(isbn string) (*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, book := range r.books {
		if book.ISBN == isbn {
			return book, nil
		}
	}
	return nil, ErrBookNotFound
}

func (r *InMemoryBookRepository) Update(book *models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[book.ID]; !exists {
		return ErrBookNotFound
	}
	r.books[book.ID] = book
	return nil
}

func (r *InMemoryBookRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return ErrBookNotFound
	}
	delete(r.books, id)
	return nil
}

func (r *InMemoryBookRepository) GetAll() ([]*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	books := make([]*models.Book, 0, len(r.books))
	for _, book := range r.books {
		books = append(books, book)
	}
	return books, nil
}

func (r *InMemoryBookRepository) UpdateStock(bookID string, delta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	book, exists := r.books[bookID]
	if !exists {
		return ErrBookNotFound
	}
	book.Stock += delta
	if book.Stock < 0 {
		book.Stock -= delta
		return errors.New("insufficient stock")
	}
	return nil
}
