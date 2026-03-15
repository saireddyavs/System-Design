package repositories

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
)

var (
	ErrFileNotFound   = errors.New("file not found")
	ErrFolderNotFound = errors.New("folder not found")
)

// InMemoryFileRepo implements FileRepository with in-memory storage.
type InMemoryFileRepo struct {
	mu      sync.RWMutex
	files   map[string]*models.File
	folders map[string]*models.Folder
}

// NewInMemoryFileRepo creates a new in-memory file repository.
func NewInMemoryFileRepo() interfaces.FileRepository {
	return &InMemoryFileRepo{
		files:   make(map[string]*models.File),
		folders: make(map[string]*models.Folder),
	}
}

func (r *InMemoryFileRepo) CreateFile(file *models.File) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.files[file.ID]; exists {
		return errors.New("file already exists")
	}
	r.files[file.ID] = file
	// Composite: add to parent folder's children
	if file.ParentFolderID != "" {
		if parent, exists := r.folders[file.ParentFolderID]; exists {
			parent.AddChild(file)
		}
	}
	return nil
}

func (r *InMemoryFileRepo) GetFileByID(id string) (*models.File, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	file, exists := r.files[id]
	if !exists {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (r *InMemoryFileRepo) UpdateFile(file *models.File) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.files[file.ID]; !exists {
		return ErrFileNotFound
	}
	r.files[file.ID] = file
	return nil
}

func (r *InMemoryFileRepo) UpdateFileParent(fileID, oldParentID, newParentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	file, exists := r.files[fileID]
	if !exists {
		return ErrFileNotFound
	}
	if oldParentID != "" {
		if parent, ok := r.folders[oldParentID]; ok {
			parent.RemoveChild(fileID)
		}
	}
	if newParentID != "" {
		if parent, ok := r.folders[newParentID]; ok {
			parent.AddChild(file)
		}
	}
	return nil
}

func (r *InMemoryFileRepo) DeleteFile(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	file, exists := r.files[id]
	if !exists {
		return ErrFileNotFound
	}
	if file.ParentFolderID != "" {
		if parent, ok := r.folders[file.ParentFolderID]; ok {
			parent.RemoveChild(id)
		}
	}
	delete(r.files, id)
	return nil
}

func (r *InMemoryFileRepo) CreateFolder(folder *models.Folder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.folders[folder.ID]; exists {
		return errors.New("folder already exists")
	}
	r.folders[folder.ID] = folder
	// Composite: add to parent folder's children
	if folder.ParentFolderID != "" {
		if parent, exists := r.folders[folder.ParentFolderID]; exists {
			parent.AddChild(folder)
		}
	}
	return nil
}

func (r *InMemoryFileRepo) GetFolderByID(id string) (*models.Folder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	folder, exists := r.folders[id]
	if !exists {
		return nil, ErrFolderNotFound
	}
	return folder, nil
}

func (r *InMemoryFileRepo) UpdateFolder(folder *models.Folder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.folders[folder.ID]; !exists {
		return ErrFolderNotFound
	}
	r.folders[folder.ID] = folder
	return nil
}

func (r *InMemoryFileRepo) UpdateFolderParent(folderID, oldParentID, newParentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	folder, exists := r.folders[folderID]
	if !exists {
		return ErrFolderNotFound
	}
	if oldParentID != "" {
		if parent, ok := r.folders[oldParentID]; ok {
			parent.RemoveChild(folderID)
		}
	}
	if newParentID != "" {
		if parent, ok := r.folders[newParentID]; ok {
			parent.AddChild(folder)
		}
	}
	return nil
}

func (r *InMemoryFileRepo) DeleteFolder(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	folder, exists := r.folders[id]
	if !exists {
		return ErrFolderNotFound
	}
	if folder.ParentFolderID != "" {
		if parent, ok := r.folders[folder.ParentFolderID]; ok {
			parent.RemoveChild(id)
		}
	}
	delete(r.folders, id)
	return nil
}
