package interfaces

import "library-management-system/internal/models"

// BookRepository defines data access for books (Repository pattern)
type BookRepository interface {
	Create(book *models.Book) error
	GetByID(id string) (*models.Book, error)
	Update(book *models.Book) error
	Delete(id string) error
	ListAll() ([]*models.Book, error)
	GetByISBN(isbn string) (*models.Book, error)
}
