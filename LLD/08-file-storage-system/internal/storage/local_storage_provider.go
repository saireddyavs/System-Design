package storage

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"os"
	"path/filepath"
	"sync"
)

var ErrPathNotFound = errors.New("path not found")

// LocalStorageProvider implements StorageProvider for local filesystem (Strategy pattern).
type LocalStorageProvider struct {
	mu       sync.RWMutex
	basePath string
	// In-memory fallback for demo when basePath is empty
	inMemory map[string][]byte
}

// NewLocalStorageProvider creates a new local storage provider.
// If basePath is empty, uses in-memory storage for testing.
func NewLocalStorageProvider(basePath string) interfaces.StorageProvider {
	if basePath == "" {
		return &LocalStorageProvider{
			inMemory: make(map[string][]byte),
		}
	}
	_ = os.MkdirAll(basePath, 0755)
	return &LocalStorageProvider{
		basePath: basePath,
	}
}

func (s *LocalStorageProvider) Upload(path string, content []byte) (string, error) {
	if s.inMemory != nil {
		s.mu.Lock()
		s.inMemory[path] = content
		s.mu.Unlock()
		return path, nil
	}
	fullPath := filepath.Join(s.basePath, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return path, os.WriteFile(fullPath, content, 0644)
}

func (s *LocalStorageProvider) Download(path string) ([]byte, error) {
	if s.inMemory != nil {
		s.mu.RLock()
		content, exists := s.inMemory[path]
		s.mu.RUnlock()
		if !exists {
			return nil, ErrPathNotFound
		}
		return content, nil
	}
	fullPath := filepath.Join(s.basePath, path)
	return os.ReadFile(fullPath)
}

func (s *LocalStorageProvider) Delete(path string) error {
	if s.inMemory != nil {
		s.mu.Lock()
		delete(s.inMemory, path)
		s.mu.Unlock()
		return nil
	}
	fullPath := filepath.Join(s.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return ErrPathNotFound
	}
	return os.Remove(fullPath)
}

func (s *LocalStorageProvider) Exists(path string) (bool, error) {
	if s.inMemory != nil {
		s.mu.RLock()
		_, exists := s.inMemory[path]
		s.mu.RUnlock()
		return exists, nil
	}
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
