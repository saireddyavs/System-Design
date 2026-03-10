package services

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"file-storage-system/internal/utils"
	"sync"
	"time"
)

var (
	ErrQuotaExceeded       = errors.New("storage quota exceeded")
	ErrFileNotFound        = errors.New("file not found")
	ErrFolderNotFound      = errors.New("folder not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrParentFolderInvalid = errors.New("parent folder not found or invalid")
)

// FileService handles file operations.
type FileService struct {
	mu              sync.RWMutex
	fileRepo        interfaces.FileRepository
	userRepo        interfaces.UserRepository
	storageProvider interfaces.StorageProvider
	versionRepo     interfaces.VersionRepository
	searchEngine    interfaces.SearchEngine
	userService     *UserService
	versionService  *VersionService
}

// NewFileService creates a new file service.
func NewFileService(
	fileRepo interfaces.FileRepository,
	userRepo interfaces.UserRepository,
	storageProvider interfaces.StorageProvider,
	versionRepo interfaces.VersionRepository,
	searchEngine interfaces.SearchEngine,
	userService *UserService,
	versionService *VersionService,
) *FileService {
	return &FileService{
		fileRepo:        fileRepo,
		userRepo:        userRepo,
		storageProvider: storageProvider,
		versionRepo:     versionRepo,
		searchEngine:    searchEngine,
		userService:     userService,
		versionService: versionService,
	}
}

// Upload creates a new file and stores its content.
func (s *FileService) Upload(userID, name, parentFolderID, mimeType string, content []byte) (*models.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check quota
	ok, err := s.userService.CheckQuota(userID, int64(len(content)))
	if err != nil || !ok {
		return nil, ErrQuotaExceeded
	}

	// Validate parent folder
	if parentFolderID != "" {
		_, err := s.fileRepo.GetFolderByID(parentFolderID)
		if err != nil {
			return nil, ErrParentFolderInvalid
		}
	}

	fileID := utils.GenerateID()
	file := models.NewFile(fileID, name, userID, parentFolderID, int64(len(content)), mimeType, content)

	// Store in storage provider
	storagePath := userID + "/" + fileID
	if _, err := s.storageProvider.Upload(storagePath, content); err != nil {
		return nil, err
	}

	if err := s.fileRepo.CreateFile(file); err != nil {
		s.storageProvider.Delete(storagePath)
		return nil, err
	}

	// Update user storage
	user, _ := s.userRepo.GetByID(userID)
	user.AddUsedStorage(int64(len(content)))

	// Index for search
	path := name
	if parentFolderID != "" {
		path = parentFolderID + "/" + name
	}
	_ = s.searchEngine.Index(file, path)

	return file, nil
}

// Download retrieves file content.
func (s *FileService) Download(userID, fileID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return nil, ErrFileNotFound
	}

	// Check permission (owner or shared with view/edit)
	if file.OwnerID != userID {
		// Would need sharing service to check - simplified for now
		// In full impl: sharingService.HasAccess(userID, fileID, PermissionView)
	}

	storagePath := file.OwnerID + "/" + fileID
	return s.storageProvider.Download(storagePath)
}

// Delete removes a file. Only owner can delete.
func (s *FileService) Delete(userID, fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if file.OwnerID != userID {
		return ErrPermissionDenied
	}

	storagePath := file.OwnerID + "/" + fileID
	_ = s.storageProvider.Delete(storagePath)
	if err := s.fileRepo.DeleteFile(fileID); err != nil {
		return err
	}
	_ = s.versionRepo.DeleteByFileID(fileID)
	_ = s.searchEngine.RemoveFromIndex(fileID)

	user, _ := s.userRepo.GetByID(userID)
	user.RemoveUsedStorage(file.Size)

	return nil
}

// Move moves a file to a new parent folder.
func (s *FileService) Move(userID, fileID, newParentFolderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if file.OwnerID != userID {
		return ErrPermissionDenied
	}

	if newParentFolderID != "" {
		_, err := s.fileRepo.GetFolderByID(newParentFolderID)
		if err != nil {
			return ErrParentFolderInvalid
		}
	}

	oldParentID := file.ParentFolderID
	file.ParentFolderID = newParentFolderID
	file.UpdatedAt = time.Now()
	if err := s.fileRepo.UpdateFile(file); err != nil {
		return err
	}
	return s.fileRepo.UpdateFileParent(fileID, oldParentID, newParentFolderID)
}

// Rename renames a file.
func (s *FileService) Rename(userID, fileID, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if file.OwnerID != userID {
		return ErrPermissionDenied
	}

	file.Name = newName
	file.UpdatedAt = time.Now()
	return s.fileRepo.UpdateFile(file)
}
