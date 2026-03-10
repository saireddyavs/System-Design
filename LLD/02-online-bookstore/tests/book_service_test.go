package tests

import (
	"testing"

	"online-bookstore/internal/repositories"
	"online-bookstore/internal/services"
)

func TestBookService_CreateBook(t *testing.T) {
	bookRepo := repositories.NewInMemoryBookRepository()
	bookSvc := services.NewBookService(bookRepo)

	book, err := bookSvc.CreateBook("Test Book", "Test Author", "978-1234567890", 29.99, "Fiction", 10)
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}
	if book.ID == "" {
		t.Error("expected non-empty ID")
	}
	if book.Title != "Test Book" {
		t.Errorf("expected title 'Test Book', got %s", book.Title)
	}
	if book.Stock != 10 {
		t.Errorf("expected stock 10, got %d", book.Stock)
	}
}

func TestBookService_CreateBook_DuplicateISBN(t *testing.T) {
	bookRepo := repositories.NewInMemoryBookRepository()
	bookSvc := services.NewBookService(bookRepo)

	_, _ = bookSvc.CreateBook("Book 1", "Author 1", "978-1111111111", 10, "Genre", 5)
	_, err := bookSvc.CreateBook("Book 2", "Author 2", "978-1111111111", 20, "Genre", 3)
	if err != services.ErrBookExists {
		t.Errorf("expected ErrBookExists, got %v", err)
	}
}

func TestBookService_GetBook(t *testing.T) {
	bookRepo := repositories.NewInMemoryBookRepository()
	bookSvc := services.NewBookService(bookRepo)

	created, _ := bookSvc.CreateBook("Get Test", "Author", "978-2222222222", 15, "Tech", 7)
	retrieved, err := bookSvc.GetBook(created.ID)
	if err != nil {
		t.Fatalf("GetBook failed: %v", err)
	}
	if retrieved.Title != created.Title {
		t.Errorf("expected %s, got %s", created.Title, retrieved.Title)
	}
}

func TestBookService_ListBooks(t *testing.T) {
	bookRepo := repositories.NewInMemoryBookRepository()
	bookSvc := services.NewBookService(bookRepo)

	_, _ = bookSvc.CreateBook("A", "A", "978-aaa", 1, "G", 1)
	_, _ = bookSvc.CreateBook("B", "B", "978-bbb", 2, "G", 2)

	books, err := bookSvc.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks failed: %v", err)
	}
	if len(books) != 2 {
		t.Errorf("expected 2 books, got %d", len(books))
	}
}
