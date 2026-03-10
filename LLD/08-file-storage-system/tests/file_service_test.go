package tests

import (
	"file-storage-system/internal/repositories"
	"file-storage-system/internal/services"
	"file-storage-system/internal/storage"
	"testing"
)

func setupFileService(t *testing.T) (*services.FileService, *services.UserService, *services.FolderService) {
	userRepo := repositories.NewInMemoryUserRepo()
	fileRepo := repositories.NewInMemoryFileRepo()
	versionRepo := repositories.NewInMemoryVersionRepo()
	searchEngine := repositories.NewInMemorySearchEngine()
	storageProvider := storage.NewLocalStorageProvider("")

	userService := services.NewUserService(userRepo)
	_, _ = userService.CreateUser("user1", "Alice", "alice@test.com")

	versionService := services.NewVersionService(fileRepo, versionRepo, userRepo)
	folderService := services.NewFolderService(fileRepo, userRepo, searchEngine)
	fileService := services.NewFileService(
		fileRepo, userRepo, storageProvider, versionRepo, searchEngine,
		userService, versionService,
	)

	return fileService, userService, folderService
}

func TestFileService_Upload(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, err := folderService.CreateFolder("user1", "Root", "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	content := []byte("test content")
	file, err := fileService.Upload("user1", "test.txt", rootFolder.ID, "text/plain", content)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if file == nil {
		t.Fatal("Expected file, got nil")
	}
	if file.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %s", file.Name)
	}
	if file.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), file.Size)
	}
	if file.OwnerID != "user1" {
		t.Errorf("Expected owner user1, got %s", file.OwnerID)
	}
}

func TestFileService_Upload_QuotaExceeded(t *testing.T) {
	fileService, userService, folderService := setupFileService(t)

	user, _ := userService.GetUser("user1")
	// Exhaust quota: 1GB default, create file larger than that
	user.StorageQuota = 10 // 10 bytes only

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	content := []byte("this is more than 10 bytes")

	_, err := fileService.Upload("user1", "large.txt", rootFolder.ID, "text/plain", content)
	if err != services.ErrQuotaExceeded {
		t.Errorf("Expected ErrQuotaExceeded, got %v", err)
	}
}

func TestFileService_Download(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	content := []byte("download me")
	file, _ := fileService.Upload("user1", "download.txt", rootFolder.ID, "text/plain", content)

	downloaded, err := fileService.Download("user1", file.ID)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if string(downloaded) != string(content) {
		t.Errorf("Expected content %q, got %q", content, downloaded)
	}
}

func TestFileService_Delete(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "todelete.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := fileService.Delete("user1", file.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = fileService.Download("user1", file.ID)
	if err != services.ErrFileNotFound {
		t.Errorf("Expected ErrFileNotFound after delete, got %v", err)
	}
}

func TestFileService_Delete_PermissionDenied(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "protected.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := fileService.Delete("user2", file.ID)
	if err != services.ErrPermissionDenied {
		t.Errorf("Expected ErrPermissionDenied, got %v", err)
	}
}

func TestFileService_Move(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	subFolder, _ := folderService.CreateFolder("user1", "Sub", rootFolder.ID)
	file, _ := fileService.Upload("user1", "move.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := fileService.Move("user1", file.ID, subFolder.ID)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}
}

func TestFileService_Rename(t *testing.T) {
	fileService, _, folderService := setupFileService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "old.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := fileService.Rename("user1", file.ID, "new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
}
