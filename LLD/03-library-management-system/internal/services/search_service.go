package services

import (
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"strings"
)

// SearchCriteria defines search filters
type SearchCriteria struct {
	Title   string
	Author  string
	Subject string
	ISBN    string
}

// SearchService handles book search (SRP)
type SearchService struct {
	bookRepo interfaces.BookRepository
}

// NewSearchService creates a new search service
func NewSearchService(bookRepo interfaces.BookRepository) *SearchService {
	return &SearchService{bookRepo: bookRepo}
}

// Search finds books matching the criteria (case-insensitive partial match)
func (s *SearchService) Search(criteria SearchCriteria) ([]*models.Book, error) {
	books, err := s.bookRepo.ListAll()
	if err != nil {
		return nil, err
	}

	var results []*models.Book
	for _, b := range books {
		if s.matches(b, criteria) {
			results = append(results, b)
		}
	}
	return results, nil
}

func (s *SearchService) matches(book *models.Book, c SearchCriteria) bool {
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
