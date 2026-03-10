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
	ErrNotOwnerCannotShare = errors.New("only owner can share")
	ErrAlreadyShared        = errors.New("already shared with user")
)

// SharingService handles file/folder sharing with Observer pattern for notifications.
type SharingService struct {
	mu               sync.RWMutex
	fileRepo         interfaces.FileRepository
	permissionRepo   interfaces.PermissionRepository
	observers        []interfaces.ShareObserver
}

// NewSharingService creates a new sharing service.
func NewSharingService(
	fileRepo interfaces.FileRepository,
	permissionRepo interfaces.PermissionRepository,
) *SharingService {
	return &SharingService{
		fileRepo:       fileRepo,
		permissionRepo: permissionRepo,
		observers:      make([]interfaces.ShareObserver, 0),
	}
}

// RegisterObserver adds an observer for share notifications (Observer pattern).
func (s *SharingService) RegisterObserver(observer interfaces.ShareObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// ShareFile shares a file with a user at the given permission level.
func (s *SharingService) ShareFile(ownerID, fileID, targetUserID string, level models.PermissionLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if file.OwnerID != ownerID {
		return ErrNotOwnerCannotShare
	}

	existing, _ := s.permissionRepo.GetByFileAndUser(fileID, targetUserID)
	if existing != nil {
		return ErrAlreadyShared
	}

	permission := &models.Permission{
		ID:        utils.GenerateID(),
		FileID:    fileID,
		UserID:    targetUserID,
		Level:     level,
		GrantedAt: time.Now(),
		GrantedBy: ownerID,
	}

	if err := s.permissionRepo.Create(permission); err != nil {
		return err
	}

	// Notify observers
	for _, obs := range s.observers {
		obs.OnFileShared(file, permission)
	}

	return nil
}

// ShareFolder shares a folder with a user (permission inheritance for children).
func (s *SharingService) ShareFolder(ownerID, folderID, targetUserID string, level models.PermissionLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.OwnerID != ownerID {
		return ErrNotOwnerCannotShare
	}

	existing, _ := s.permissionRepo.GetByFileAndUser(folderID, targetUserID)
	if existing != nil {
		return ErrAlreadyShared
	}

	permission := &models.Permission{
		ID:        utils.GenerateID(),
		FileID:    folderID,
		UserID:    targetUserID,
		Level:     level,
		GrantedAt: time.Now(),
		GrantedBy: ownerID,
	}

	if err := s.permissionRepo.Create(permission); err != nil {
		return err
	}

	for _, obs := range s.observers {
		obs.OnFolderShared(folder, permission)
	}

	return nil
}

// RevokeAccess removes sharing for a user.
func (s *SharingService) RevokeAccess(ownerID, fileID, targetUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		folder, fErr := s.fileRepo.GetFolderByID(fileID)
		if fErr != nil {
			return ErrFileNotFound
		}
		if folder.OwnerID != ownerID {
			return ErrNotOwnerCannotShare
		}
	} else {
		if file.OwnerID != ownerID {
			return ErrNotOwnerCannotShare
		}
	}

	return s.permissionRepo.DeleteByFileAndUser(fileID, targetUserID)
}

// HasAccess checks if user has at least view access.
func (s *SharingService) HasAccess(userID, fileID string, requireEdit bool) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		folder, fErr := s.fileRepo.GetFolderByID(fileID)
		if fErr != nil {
			return false, ErrFileNotFound
		}
		if folder.OwnerID == userID {
			return true, nil
		}
		perm, _ := s.permissionRepo.GetByFileAndUser(fileID, userID)
		if perm == nil {
			return false, nil
		}
		if requireEdit {
			return perm.CanEdit(), nil
		}
		return perm.CanView(), nil
	}

	if file.OwnerID == userID {
		return true, nil
	}
	perm, _ := s.permissionRepo.GetByFileAndUser(fileID, userID)
	if perm == nil {
		return false, nil
	}
	if requireEdit {
		return perm.CanEdit(), nil
	}
	return perm.CanView(), nil
}
