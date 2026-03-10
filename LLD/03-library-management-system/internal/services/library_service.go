package services

import (
	"errors"
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
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

// RemoveBook removes a book by ID (only if no copies are checked out)
func (s *LibraryService) RemoveBook(id string) error {
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		return err
	}
	if book.AvailableCopies != book.TotalCopies {
		return errors.New("cannot remove book: copies are still checked out")
	}
	return s.bookRepo.Delete(id)
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

// UpdateMember updates member details
func (s *LibraryService) UpdateMember(id, name, email string) (*models.Member, error) {
	member, err := s.memberRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	member.Name = name
	member.Email = email
	if err := s.memberRepo.Update(member); err != nil {
		return nil, err
	}
	return member, nil
}

// DeactivateMember deactivates a member
func (s *LibraryService) DeactivateMember(id string) error {
	member, err := s.memberRepo.GetByID(id)
	if err != nil {
		return err
	}
	if member.BorrowedCount > 0 {
		return errors.New("cannot deactivate: member has active loans")
	}
	member.IsActive = false
	return s.memberRepo.Update(member)
}
