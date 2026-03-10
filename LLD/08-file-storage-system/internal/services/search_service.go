package services

import (
	"file-storage-system/internal/interfaces"
	"sync"
)

// SearchService handles file search operations.
type SearchService struct {
	mu           sync.RWMutex
	searchEngine interfaces.SearchEngine
}

// NewSearchService creates a new search service.
func NewSearchService(searchEngine interfaces.SearchEngine) *SearchService {
	return &SearchService{
		searchEngine: searchEngine,
	}
}

// SearchByName searches files and folders by name for a user.
func (s *SearchService) SearchByName(userID, query string) ([]*interfaces.SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchEngine.SearchByName(userID, query)
}
