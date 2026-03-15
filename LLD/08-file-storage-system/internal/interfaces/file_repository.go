package interfaces

import "file-storage-system/internal/models"

// FileRepository defines the contract for file and folder data access (Repository pattern).
type FileRepository interface {
	// File operations
	CreateFile(file *models.File) error
	GetFileByID(id string) (*models.File, error)
	UpdateFile(file *models.File) error
	UpdateFileParent(fileID, oldParentID, newParentID string) error
	DeleteFile(id string) error

	// Folder operations
	CreateFolder(folder *models.Folder) error
	GetFolderByID(id string) (*models.Folder, error)
	UpdateFolder(folder *models.Folder) error
	UpdateFolderParent(folderID, oldParentID, newParentID string) error
	DeleteFolder(id string) error
}
