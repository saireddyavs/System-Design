package interfaces

// StorageProvider defines the storage strategy (Strategy pattern).
// Different implementations: LocalStorage, S3Storage, GCSStorage.
type StorageProvider interface {
	// Upload stores content and returns the storage path.
	Upload(path string, content []byte) (string, error)

	// Download retrieves content from storage path.
	Download(path string) ([]byte, error)

	// Delete removes content from storage.
	Delete(path string) error
}
