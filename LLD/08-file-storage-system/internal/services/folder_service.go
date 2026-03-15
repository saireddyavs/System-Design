package services

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"

	"file-storage-system/internal/utils"
)

var (
	ErrFolderAlreadyExists = errors.New("folder already exists")
	ErrCannotDeleteNonEmpty = errors.New("cannot delete non-empty folder")
)

// FolderService handles folder operations (Composite pattern).
type FolderService struct {
	mu          sync.RWMutex
	fileRepo    interfaces.FileRepository
	userRepo    interfaces.UserRepository
	searchEngine interfaces.SearchEngine
}

// NewFolderService creates a new folder service.
func NewFolderService(
	fileRepo interfaces.FileRepository,
	userRepo interfaces.UserRepository,
	searchEngine interfaces.SearchEngine,
) *FolderService {
	return &FolderService{
		fileRepo:    fileRepo,
		userRepo:    userRepo,
		searchEngine: searchEngine,
	}
}

// CreateFolder creates a new folder (Factory pattern - creates folder type).
func (s *FolderService) CreateFolder(userID, name, parentFolderID string) (*models.Folder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if parentFolderID != "" {
		_, err := s.fileRepo.GetFolderByID(parentFolderID)
		if err != nil {
			return nil, ErrParentFolderInvalid
		}
	}

	folderID := utils.GenerateID()
	folder := models.NewFolder(folderID, name, userID, parentFolderID)

	if err := s.fileRepo.CreateFolder(folder); err != nil {
		return nil, err
	}

	path := name
	if parentFolderID != "" {
		path = parentFolderID + "/" + name
	}
	_ = s.searchEngine.Index(folder, path)

	return folder, nil
}

// DeleteFolder deletes a folder. Only empty folders or recursive delete.
func (s *FolderService) DeleteFolder(userID, folderID string, recursive bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.OwnerID != userID {
		return ErrPermissionDenied
	}

	if !recursive && len(folder.GetChildren()) > 0 {
		return ErrCannotDeleteNonEmpty
	}

	// Recursive delete would remove all children - simplified for LLD
	if err := s.fileRepo.DeleteFolder(folderID); err != nil {
		return err
	}
	_ = s.searchEngine.RemoveFromIndex(folderID)
	return nil
}

// GetFolderSize returns total size of folder (Composite - recursive sum).
func (s *FolderService) GetFolderSize(folderID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil {
		return 0, ErrFolderNotFound
	}
	return folder.GetSize(), nil
}
