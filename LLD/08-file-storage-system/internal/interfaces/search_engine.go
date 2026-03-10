package interfaces

import "file-storage-system/internal/models"

// SearchResult represents a search hit.
type SearchResult struct {
	Item     models.FileSystemItem
	IsFile   bool
	File     *models.File
	Folder   *models.Folder
	Path     string
}

// SearchEngine defines the contract for file search (Strategy pattern).
type SearchEngine interface {
	// SearchByName searches files and folders by name.
	SearchByName(userID string, query string) ([]*SearchResult, error)

	// Index adds or updates an item in the search index.
	Index(item models.FileSystemItem, path string) error

	// RemoveFromIndex removes an item from the search index.
	RemoveFromIndex(itemID string) error
}
