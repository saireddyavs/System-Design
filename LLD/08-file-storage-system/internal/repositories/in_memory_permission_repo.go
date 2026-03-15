package repositories

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
)

var ErrPermissionNotFound = errors.New("permission not found")

// InMemoryPermissionRepo implements PermissionRepository.
type InMemoryPermissionRepo struct {
	mu          sync.RWMutex
	permissions map[string]*models.Permission
	byFile      map[string][]string // fileID -> permission IDs
	byUser      map[string][]string // userID -> permission IDs
}

// NewInMemoryPermissionRepo creates a new in-memory permission repository.
func NewInMemoryPermissionRepo() interfaces.PermissionRepository {
	return &InMemoryPermissionRepo{
		permissions: make(map[string]*models.Permission),
		byFile:      make(map[string][]string),
		byUser:      make(map[string][]string),
	}
}

func (r *InMemoryPermissionRepo) Create(permission *models.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.permissions[permission.ID]; exists {
		return errors.New("permission already exists")
	}
	r.permissions[permission.ID] = permission
	r.byFile[permission.FileID] = append(r.byFile[permission.FileID], permission.ID)
	r.byUser[permission.UserID] = append(r.byUser[permission.UserID], permission.ID)
	return nil
}

func (r *InMemoryPermissionRepo) GetByFileID(fileID string) ([]*models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byFile[fileID]
	var result []*models.Permission
	for _, id := range ids {
		if p, exists := r.permissions[id]; exists {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryPermissionRepo) GetByFileAndUser(fileID, userID string) (*models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byFile[fileID]
	for _, id := range ids {
		if p, exists := r.permissions[id]; exists && p.UserID == userID {
			return p, nil
		}
	}
	return nil, ErrPermissionNotFound
}

func (r *InMemoryPermissionRepo) DeleteByFileAndUser(fileID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.byFile[fileID]
	for _, id := range ids {
		if p, exists := r.permissions[id]; exists && p.UserID == userID {
			delete(r.permissions, id)
			r.removeFromIndex(fileID, userID, id)
			return nil
		}
	}
	return ErrPermissionNotFound
}

func (r *InMemoryPermissionRepo) removeFromIndex(fileID, userID, permID string) {
	// Remove from byFile
	if ids, ok := r.byFile[fileID]; ok {
		for i, id := range ids {
			if id == permID {
				r.byFile[fileID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	// Remove from byUser
	if ids, ok := r.byUser[userID]; ok {
		for i, id := range ids {
			if id == permID {
				r.byUser[userID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
}
