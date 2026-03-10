package interfaces

import "file-storage-system/internal/models"

// VersionRepository defines the contract for version data access.
type VersionRepository interface {
	Create(version *models.Version) error
	GetByFileID(fileID string) ([]*models.Version, error)
	GetByFileAndVersion(fileID string, versionNumber int) (*models.Version, error)
	DeleteByFileID(fileID string) error
	GetLatestVersion(fileID string) (*models.Version, error)
}
