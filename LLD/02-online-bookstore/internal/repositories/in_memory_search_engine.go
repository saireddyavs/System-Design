package repositories

import (
	"strings"
	"sync"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

// InMemorySearchEngine implements SearchEngine with efficient case-insensitive substring matching.
// Uses strings.Contains with strings.ToLower for O(n*m) substring search per book.
// For production: replace with Elasticsearch/Meilisearch for O(log n) full-text search.
type InMemorySearchEngine struct {
	bookRepo interfaces.BookRepository
	mu       sync.RWMutex
}

// NewInMemorySearchEngine creates a search engine backed by book repository.
func NewInMemorySearchEngine(bookRepo interfaces.BookRepository) *InMemorySearchEngine {
	return &InMemorySearchEngine{bookRepo: bookRepo}
}

// containsIgnoreCase performs case-insensitive substring matching.
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func (s *InMemorySearchEngine) SearchByTitle(query string) ([]*models.Book, error) {
	return s.search(query, func(book *models.Book) bool {
		return containsIgnoreCase(book.Title, query)
	})
}

func (s *InMemorySearchEngine) SearchByAuthor(query string) ([]*models.Book, error) {
	return s.search(query, func(book *models.Book) bool {
		return containsIgnoreCase(book.Author, query)
	})
}

func (s *InMemorySearchEngine) SearchByGenre(query string) ([]*models.Book, error) {
	return s.search(query, func(book *models.Book) bool {
		return containsIgnoreCase(book.Genre, query)
	})
}

func (s *InMemorySearchEngine) Search(query string, searchType interfaces.SearchType) ([]*models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	books, err := s.bookRepo.GetAll()
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return books, nil
	}

	var result []*models.Book
	for _, book := range books {
		matched := false
		switch searchType {
		case interfaces.SearchTypeTitle:
			matched = containsIgnoreCase(book.Title, query)
		case interfaces.SearchTypeAuthor:
			matched = containsIgnoreCase(book.Author, query)
		case interfaces.SearchTypeGenre:
			matched = containsIgnoreCase(book.Genre, query)
		case interfaces.SearchTypeAll:
			matched = containsIgnoreCase(book.Title, query) ||
				containsIgnoreCase(book.Author, query) ||
				containsIgnoreCase(book.Genre, query)
		}
		if matched {
			result = append(result, book)
		}
	}
	return result, nil
}

func (s *InMemorySearchEngine) search(query string, matchFn func(*models.Book) bool) ([]*models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	books, err := s.bookRepo.GetAll()
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return books, nil
	}

	var result []*models.Book
	for _, book := range books {
		if matchFn(book) {
			result = append(result, book)
		}
	}
	return result, nil
}
