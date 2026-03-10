package services

import (
	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

// SearchService provides book search functionality (SRP).
// Delegates to SearchEngine for efficient searching.
type SearchService struct {
	searchEngine interfaces.SearchEngine
}

func NewSearchService(searchEngine interfaces.SearchEngine) *SearchService {
	return &SearchService{searchEngine: searchEngine}
}

func (s *SearchService) SearchByTitle(query string) ([]*models.Book, error) {
	return s.searchEngine.SearchByTitle(query)
}

func (s *SearchService) SearchByAuthor(query string) ([]*models.Book, error) {
	return s.searchEngine.SearchByAuthor(query)
}

func (s *SearchService) SearchByGenre(query string) ([]*models.Book, error) {
	return s.searchEngine.SearchByGenre(query)
}

func (s *SearchService) Search(query string, searchType interfaces.SearchType) ([]*models.Book, error) {
	return s.searchEngine.Search(query, searchType)
}
