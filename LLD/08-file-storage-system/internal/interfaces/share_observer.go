package interfaces

import "file-storage-system/internal/models"

// ShareObserver defines the Observer interface for sharing notifications.
// When a file is shared, all observers are notified.
type ShareObserver interface {
	OnFileShared(file *models.File, permission *models.Permission)
	OnFolderShared(folder *models.Folder, permission *models.Permission)
}
