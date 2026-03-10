package interfaces

import "online-bookstore/internal/models"

// SearchEngine defines the contract for book search operations.
// OCP: New search algorithms can be added without modifying existing code.
// Enables efficient searching with different backends (in-memory, Elasticsearch, etc.)
type SearchEngine interface {
	SearchByTitle(query string) ([]*models.Book, error)
	SearchByAuthor(query string) ([]*models.Book, error)
	SearchByGenre(query string) ([]*models.Book, error)
	Search(query string, searchType SearchType) ([]*models.Book, error)
}

// SearchType specifies which field to search.
type SearchType int

const (
	SearchTypeTitle SearchType = iota
	SearchTypeAuthor
	SearchTypeGenre
	SearchTypeAll
)
