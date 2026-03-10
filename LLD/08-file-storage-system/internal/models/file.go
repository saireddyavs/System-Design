package models

import (
	"sync"
	"time"
)

// File represents a file in the storage system.
type File struct {
	mu            sync.RWMutex
	ID            string
	Name          string
	OwnerID       string
	ParentFolderID string
	Size          int64
	MimeType      string
	Content       []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Versions      []*Version
}

// NewFile creates a new file instance.
func NewFile(id, name, ownerID, parentFolderID string, size int64, mimeType string, content []byte) *File {
	now := time.Now()
	return &File{
		ID:             id,
		Name:           name,
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
		Size:           size,
		MimeType:       mimeType,
		Content:        content,
		CreatedAt:      now,
		UpdatedAt:      now,
		Versions:       make([]*Version, 0),
	}
}

// GetID implements FileSystemItem.
func (f *File) GetID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.ID
}

// GetName implements FileSystemItem.
func (f *File) GetName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Name
}

// GetOwnerID implements FileSystemItem.
func (f *File) GetOwnerID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.OwnerID
}

// GetParentFolderID implements FileSystemItem.
func (f *File) GetParentFolderID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.ParentFolderID
}

// GetSize implements FileSystemItem.
func (f *File) GetSize() int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Size
}

// GetCreatedAt implements FileSystemItem.
func (f *File) GetCreatedAt() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.CreatedAt
}

// IsFolder implements FileSystemItem.
func (f *File) IsFolder() bool {
	return false
}

// AddVersion adds a version to the file history.
func (f *File) AddVersion(v *Version) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Versions = append(f.Versions, v)
}

// GetVersions returns a copy of versions.
func (f *File) GetVersions() []*Version {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]*Version, len(f.Versions))
	copy(result, f.Versions)
	return result
}
