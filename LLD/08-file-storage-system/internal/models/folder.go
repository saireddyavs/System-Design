package models

import (
	"sync"
	"time"
)

// FileSystemItem is the composite interface - both File and Folder implement this.
// This enables recursive operations like size calculation, deletion, permission inheritance.
type FileSystemItem interface {
	GetID() string
	GetName() string
	GetOwnerID() string
	GetParentFolderID() string
	GetSize() int64
	GetCreatedAt() time.Time
	IsFolder() bool
}

// Folder represents a folder/directory in the storage system.
// Implements FileSystemItem for the Composite pattern.
type Folder struct {
	mu            sync.RWMutex
	ID            string
	Name          string
	OwnerID       string
	ParentFolderID string
	Children      []FileSystemItem // Can contain Files and sub-Folders
	CreatedAt     time.Time
}

// NewFolder creates a new folder instance.
func NewFolder(id, name, ownerID, parentFolderID string) *Folder {
	return &Folder{
		ID:             id,
		Name:           name,
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
		Children:       make([]FileSystemItem, 0),
		CreatedAt:      time.Now(),
	}
}

// GetID implements FileSystemItem.
func (f *Folder) GetID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.ID
}

// GetName implements FileSystemItem.
func (f *Folder) GetName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Name
}

// GetOwnerID implements FileSystemItem.
func (f *Folder) GetOwnerID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.OwnerID
}

// GetParentFolderID implements FileSystemItem.
func (f *Folder) GetParentFolderID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.ParentFolderID
}

// GetSize implements FileSystemItem - recursively sums children sizes.
func (f *Folder) GetSize() int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var total int64
	for _, child := range f.Children {
		total += child.GetSize()
	}
	return total
}

// GetCreatedAt implements FileSystemItem.
func (f *Folder) GetCreatedAt() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.CreatedAt
}

// IsFolder implements FileSystemItem.
func (f *Folder) IsFolder() bool {
	return true
}

// AddChild adds a child (file or folder) to this folder.
func (f *Folder) AddChild(item FileSystemItem) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Children = append(f.Children, item)
}

// RemoveChild removes a child by ID.
func (f *Folder) RemoveChild(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, child := range f.Children {
		if child.GetID() == id {
			f.Children = append(f.Children[:i], f.Children[i+1:]...)
			return true
		}
	}
	return false
}

// GetChildren returns a copy of children.
func (f *Folder) GetChildren() []FileSystemItem {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]FileSystemItem, len(f.Children))
	copy(result, f.Children)
	return result
}
