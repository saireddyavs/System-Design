package interfaces

import "online-bookstore/internal/models"

// BookRepository defines the contract for book data access.
// ISP: Focused interface - only book-related operations.
// LSP: Any implementation can substitute this interface.
type BookRepository interface {
	Create(book *models.Book) error
	GetByID(id string) (*models.Book, error)
	GetByISBN(isbn string) (*models.Book, error)
	Update(book *models.Book) error
	Delete(id string) error
	GetAll() ([]*models.Book, error)
	UpdateStock(bookID string, delta int) error
}
