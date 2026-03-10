package repositories

import (
	"errors"
	"library-management-system/internal/models"
	"library-management-system/internal/interfaces"
	"sync"
)

var ErrBookNotFound = errors.New("book not found")

// InMemoryBookRepo implements BookRepository with thread-safe in-memory storage
type InMemoryBookRepo struct {
	books map[string]*models.Book
	mu    sync.RWMutex
}

// NewInMemoryBookRepo creates a new in-memory book repository
func NewInMemoryBookRepo() interfaces.BookRepository {
	return &InMemoryBookRepo{
		books: make(map[string]*models.Book),
	}
}

func (r *InMemoryBookRepo) Create(book *models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.books[book.ID] = book
	return nil
}

func (r *InMemoryBookRepo) GetByID(id string) (*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	book, ok := r.books[id]
	if !ok {
		return nil, ErrBookNotFound
	}
	return book, nil
}

func (r *InMemoryBookRepo) Update(book *models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[book.ID]; !ok {
		return ErrBookNotFound
	}
	r.books[book.ID] = book
	return nil
}

func (r *InMemoryBookRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.books, id)
	return nil
}

func (r *InMemoryBookRepo) ListAll() ([]*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	books := make([]*models.Book, 0, len(r.books))
	for _, b := range r.books {
		books = append(books, b)
	}
	return books, nil
}

func (r *InMemoryBookRepo) GetByISBN(isbn string) (*models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.books {
		if b.ISBN == isbn {
			return b, nil
		}
	}
	return nil, ErrBookNotFound
}
