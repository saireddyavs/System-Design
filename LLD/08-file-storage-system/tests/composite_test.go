package tests

import (
	"file-storage-system/internal/repositories"
	"file-storage-system/internal/services"
	"file-storage-system/internal/storage"
	"testing"
)

func TestComposite_FolderSize(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepo()
	fileRepo := repositories.NewInMemoryFileRepo()
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

	// Create hierarchy: root -> subfolder -> file
	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	subFolder, _ := folderService.CreateFolder("user1", "Sub", rootFolder.ID)
	_, _ = fileService.Upload("user1", "a.txt", rootFolder.ID, "text/plain", []byte("aaa"))   // 3 bytes
	_, _ = fileService.Upload("user1", "b.txt", subFolder.ID, "text/plain", []byte("bbbb")) // 4 bytes

	// Composite: folder size = sum of children
	size, err := folderService.GetFolderSize(rootFolder.ID)
	if err != nil {
		t.Fatalf("GetFolderSize failed: %v", err)
	}
	// Root contains: Sub (folder with b.txt=4) + a.txt (3) = 7
	if size != 7 {
		t.Errorf("Expected folder size 7 (3+4), got %d", size)
	}
}
