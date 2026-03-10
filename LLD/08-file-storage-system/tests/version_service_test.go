package tests

import (
	"file-storage-system/internal/repositories"
	"file-storage-system/internal/services"
	"file-storage-system/internal/storage"
	"testing"
)

func setupVersionService(t *testing.T) (*services.VersionService, *services.FileService, *services.FolderService) {
	userRepo := repositories.NewInMemoryUserRepo()
	fileRepo := repositories.NewInMemoryFileRepo()
	permissionRepo := repositories.NewInMemoryPermissionRepo()
	versionRepo := repositories.NewInMemoryVersionRepo()
	searchEngine := repositories.NewInMemorySearchEngine()
	storageProvider := storage.NewLocalStorageProvider("")

	userService := services.NewUserService(userRepo)
	_, _ = userService.CreateUser("user1", "Alice", "alice@test.com")

	versionService := services.NewVersionService(fileRepo, versionRepo, userRepo)
	fileService := services.NewFileService(
		fileRepo, userRepo, storageProvider, versionRepo, searchEngine,
		userService, versionService,
	)
	folderService := services.NewFolderService(fileRepo, userRepo, searchEngine)

	_ = permissionRepo
	return versionService, fileService, folderService
}

func TestVersionService_CreateVersion(t *testing.T) {
	versionService, fileService, folderService := setupVersionService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "versioned.txt", rootFolder.ID, "text/plain", []byte("v1"))

	// Create version (simulating update)
	version, err := versionService.CreateVersion(file.ID, "user1", []byte("v2 content"))
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}
	if version.VersionNumber != 1 {
		t.Errorf("Expected version 1, got %d", version.VersionNumber)
	}
	if string(version.Content) != "v2 content" {
		t.Errorf("Expected content 'v2 content', got %s", version.Content)
	}
}

func TestVersionService_GetVersionHistory(t *testing.T) {
	versionService, fileService, folderService := setupVersionService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "history.txt", rootFolder.ID, "text/plain", []byte("v1"))
	_, _ = versionService.CreateVersion(file.ID, "user1", []byte("v2"))
	_, _ = versionService.CreateVersion(file.ID, "user1", []byte("v3"))

	versions, err := versionService.GetVersionHistory(file.ID)
	if err != nil {
		t.Fatalf("GetVersionHistory failed: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions (v2, v3), got %d", len(versions))
	}
}

func TestVersionService_RestoreVersion(t *testing.T) {
	versionService, fileService, folderService := setupVersionService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "restore.txt", rootFolder.ID, "text/plain", []byte("original"))
	_, _ = versionService.CreateVersion(file.ID, "user1", []byte("updated"))

	err := versionService.RestoreVersion("user1", file.ID, 1)
	if err != nil {
		t.Fatalf("RestoreVersion failed: %v", err)
	}
}
