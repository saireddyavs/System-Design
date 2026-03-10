package repositories

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
)

var ErrVersionNotFound = errors.New("version not found")

const MaxVersionsPerFile = 10

// InMemoryVersionRepo implements VersionRepository.
type InMemoryVersionRepo struct {
	mu        sync.RWMutex
	versions  map[string]*models.Version
	byFileID  map[string][]string // fileID -> version IDs (ordered)
}

// NewInMemoryVersionRepo creates a new in-memory version repository.
func NewInMemoryVersionRepo() interfaces.VersionRepository {
	return &InMemoryVersionRepo{
		versions: make(map[string]*models.Version),
		byFileID: make(map[string][]string),
	}
}

func (r *InMemoryVersionRepo) Create(version *models.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.versions[version.ID]; exists {
		return errors.New("version already exists")
	}
	r.versions[version.ID] = version
	r.byFileID[version.FileID] = append(r.byFileID[version.FileID], version.ID)

	// Enforce max versions: keep last 10
	ids := r.byFileID[version.FileID]
	if len(ids) > MaxVersionsPerFile {
		toRemove := ids[:len(ids)-MaxVersionsPerFile]
		r.byFileID[version.FileID] = ids[len(ids)-MaxVersionsPerFile:]
		for _, id := range toRemove {
			delete(r.versions, id)
		}
	}
	return nil
}

func (r *InMemoryVersionRepo) GetByFileID(fileID string) ([]*models.Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byFileID[fileID]
	var result []*models.Version
	for _, id := range ids {
		if v, exists := r.versions[id]; exists {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *InMemoryVersionRepo) GetByFileAndVersion(fileID string, versionNumber int) (*models.Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byFileID[fileID]
	for _, id := range ids {
		if v, exists := r.versions[id]; exists && v.VersionNumber == versionNumber {
			return v, nil
		}
	}
	return nil, ErrVersionNotFound
}

func (r *InMemoryVersionRepo) DeleteByFileID(fileID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.byFileID[fileID]
	for _, id := range ids {
		delete(r.versions, id)
	}
	delete(r.byFileID, fileID)
	return nil
}

func (r *InMemoryVersionRepo) GetLatestVersion(fileID string) (*models.Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byFileID[fileID]
	if len(ids) == 0 {
		return nil, ErrVersionNotFound
	}
	lastID := ids[len(ids)-1]
	if v, exists := r.versions[lastID]; exists {
		return v, nil
	}
	return nil, ErrVersionNotFound
}
