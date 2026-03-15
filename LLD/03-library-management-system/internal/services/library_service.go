package services

import (
	"errors"
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBookAlreadyExists = errors.New("book with this ISBN already exists")
	ErrMemberAlreadyExists = errors.New("member with this email already exists")
)

// LibraryService handles book and member management (SRP)
type LibraryService struct {
	bookRepo   interfaces.BookRepository
	memberRepo interfaces.MemberRepository
}

// NewLibraryService creates a new library service
func NewLibraryService(bookRepo interfaces.BookRepository, memberRepo interfaces.MemberRepository) *LibraryService {
	return &LibraryService{
		bookRepo:   bookRepo,
		memberRepo: memberRepo,
	}
}

// AddBook adds a new book with specified copies
func (s *LibraryService) AddBook(title, author, isbn, subject string, totalCopies int) (*models.Book, error) {
	existing, _ := s.bookRepo.GetByISBN(isbn)
	if existing != nil {
		return nil, ErrBookAlreadyExists
	}

	now := time.Now()
	book := &models.Book{
		ID:              uuid.New().String(),
		Title:           title,
		Author:          author,
		ISBN:            isbn,
		Subject:         subject,
		TotalCopies:     totalCopies,
		AvailableCopies: totalCopies,
		Status:          models.BookStatusAvailable,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}
	return book, nil
}

// RegisterMember registers a new member
func (s *LibraryService) RegisterMember(name, email string, membershipType models.MembershipType) (*models.Member, error) {
	existing, _ := s.memberRepo.GetByEmail(email)
	if existing != nil {
		return nil, ErrMemberAlreadyExists
	}

	now := time.Now()
	member := &models.Member{
		ID:               uuid.New().String(),
		Name:             name,
		Email:            email,
		MembershipType:   membershipType,
		JoinDate:         now,
		IsActive:         true,
		BorrowedCount:    0,
		MaxBorrowedLimit: models.DefaultMaxBorrowed(membershipType),
	}
	if err := s.memberRepo.Create(member); err != nil {
		return nil, err
	}
	return member, nil
}

// SearchCriteria defines search filters
type SearchCriteria struct {
	Title   string
	Author  string
	Subject string
	ISBN    string
}

// SearchBooks finds books matching the criteria (case-insensitive partial match)
func (s *LibraryService) SearchBooks(criteria SearchCriteria) ([]*models.Book, error) {
	books, err := s.bookRepo.ListAll()
	if err != nil {
		return nil, err
	}

	var results []*models.Book
	for _, b := range books {
		if s.matchesBook(b, criteria) {
			results = append(results, b)
		}
	}
	return results, nil
}

func (s *LibraryService) matchesBook(book *models.Book, c SearchCriteria) bool {
	if c.ISBN != "" && !strings.EqualFold(book.ISBN, c.ISBN) {
		return false
	}
	if c.Title != "" && !strings.Contains(strings.ToLower(book.Title), strings.ToLower(c.Title)) {
		return false
	}
	if c.Author != "" && !strings.Contains(strings.ToLower(book.Author), strings.ToLower(c.Author)) {
		return false
	}
	if c.Subject != "" && !strings.Contains(strings.ToLower(book.Subject), strings.ToLower(c.Subject)) {
		return false
	}
	return true
}
