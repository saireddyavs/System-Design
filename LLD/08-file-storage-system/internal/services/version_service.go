package services

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
	"time"

	"file-storage-system/internal/utils"
)

var (
	ErrNoVersionsAvailable = errors.New("no versions available")
	ErrVersionNotFound     = errors.New("version not found")
)

const MaxVersions = 10

// VersionService handles file versioning.
type VersionService struct {
	mu           sync.RWMutex
	fileRepo     interfaces.FileRepository
	versionRepo  interfaces.VersionRepository
	userRepo     interfaces.UserRepository
}

// NewVersionService creates a new version service.
func NewVersionService(
	fileRepo interfaces.FileRepository,
	versionRepo interfaces.VersionRepository,
	userRepo interfaces.UserRepository,
) *VersionService {
	return &VersionService{
		fileRepo:    fileRepo,
		versionRepo: versionRepo,
		userRepo:    userRepo,
	}
}

// CreateVersion creates a new version when file is updated.
func (s *VersionService) CreateVersion(fileID, userID string, content []byte) (*models.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return nil, ErrFileNotFound
	}

	versions, _ := s.versionRepo.GetByFileID(fileID)
	nextVersion := len(versions) + 1

	version := &models.Version{
		ID:            utils.GenerateID(),
		FileID:        fileID,
		VersionNumber: nextVersion,
		Content:       content,
		Size:          int64(len(content)),
		CreatedAt:     time.Now(),
		CreatedBy:     userID,
	}

	if err := s.versionRepo.Create(version); err != nil {
		return nil, err
	}

	file.AddVersion(version)
	_ = s.fileRepo.UpdateFile(file)

	return version, nil
}

// GetVersion retrieves a specific version.
func (s *VersionService) GetVersion(fileID string, versionNumber int) (*models.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.versionRepo.GetByFileAndVersion(fileID, versionNumber)
}

// GetVersionHistory returns all versions for a file.
func (s *VersionService) GetVersionHistory(fileID string) ([]*models.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.versionRepo.GetByFileID(fileID)
}

// RestoreVersion restores file content to a previous version.
func (s *VersionService) RestoreVersion(userID, fileID string, versionNumber int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if file.OwnerID != userID {
		return ErrPermissionDenied
	}

	version, err := s.versionRepo.GetByFileAndVersion(fileID, versionNumber)
	if err != nil {
		return ErrVersionNotFound
	}

	// Update file content
	file.Content = version.Content
	file.Size = version.Size
	file.UpdatedAt = time.Now()
	return s.fileRepo.UpdateFile(file)
}
