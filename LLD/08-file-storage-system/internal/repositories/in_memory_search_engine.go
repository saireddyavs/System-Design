package repositories

import (
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"strings"
	"sync"
)

// SearchIndexEntry holds indexed item with path.
type SearchIndexEntry struct {
	Item models.FileSystemItem
	Path string
	File *models.File
}

// InMemorySearchEngine implements SearchEngine with simple name matching.
type InMemorySearchEngine struct {
	mu     sync.RWMutex
	index  map[string]*SearchIndexEntry // itemID -> entry
	byUser map[string][]string          // userID -> item IDs (for filtering by access)
}

// NewInMemorySearchEngine creates a new in-memory search engine.
func NewInMemorySearchEngine() interfaces.SearchEngine {
	return &InMemorySearchEngine{
		index:  make(map[string]*SearchIndexEntry),
		byUser: make(map[string][]string),
	}
}

func (s *InMemorySearchEngine) SearchByName(userID string, query string) ([]*interfaces.SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, nil
	}
	var results []*interfaces.SearchResult
	userItems := s.byUser[userID]
	userSet := make(map[string]bool)
	for _, id := range userItems {
		userSet[id] = true
	}
	for _, entry := range s.index {
		if userID != "" && !userSet[entry.Item.GetID()] && entry.Item.GetOwnerID() != userID {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Item.GetName()), query) {
			res := &interfaces.SearchResult{
				Item:   entry.Item,
				Path:   entry.Path,
				Folder: nil,
				File:   nil,
			}
			if entry.File != nil {
				res.IsFile = true
				res.File = entry.File
			} else if folder, ok := entry.Item.(*models.Folder); ok {
				res.Folder = folder
			}
			results = append(results, res)
		}
	}
	return results, nil
}

func (s *InMemorySearchEngine) Index(item models.FileSystemItem, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := &SearchIndexEntry{Item: item, Path: path}
	if file, ok := item.(*models.File); ok {
		entry.File = file
	}
	s.index[item.GetID()] = entry
	ownerID := item.GetOwnerID()
	s.byUser[ownerID] = append(s.byUser[ownerID], item.GetID())
	return nil
}

func (s *InMemorySearchEngine) RemoveFromIndex(itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, exists := s.index[itemID]; exists {
		delete(s.index, itemID)
		ownerID := entry.Item.GetOwnerID()
		if ids, ok := s.byUser[ownerID]; ok {
			for i, id := range ids {
				if id == itemID {
					s.byUser[ownerID] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}
	return nil
}
