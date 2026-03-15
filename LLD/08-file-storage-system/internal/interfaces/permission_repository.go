package interfaces

import "file-storage-system/internal/models"

// PermissionRepository defines the contract for permission data access.
type PermissionRepository interface {
	Create(permission *models.Permission) error
	GetByFileID(fileID string) ([]*models.Permission, error)
	GetByFileAndUser(fileID, userID string) (*models.Permission, error)
	DeleteByFileAndUser(fileID, userID string) error
}
