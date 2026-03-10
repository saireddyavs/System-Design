package tests

import (
	"file-storage-system/internal/models"
	"file-storage-system/internal/repositories"
	"file-storage-system/internal/services"
	"file-storage-system/internal/storage"
	"testing"
)

func setupSharingService(t *testing.T) (*services.SharingService, *services.FileService, *services.FolderService) {
	userRepo := repositories.NewInMemoryUserRepo()
	fileRepo := repositories.NewInMemoryFileRepo()
	permissionRepo := repositories.NewInMemoryPermissionRepo()
	versionRepo := repositories.NewInMemoryVersionRepo()
	searchEngine := repositories.NewInMemorySearchEngine()
	storageProvider := storage.NewLocalStorageProvider("")

	userService := services.NewUserService(userRepo)
	_, _ = userService.CreateUser("user1", "Alice", "alice@test.com")
	_, _ = userService.CreateUser("user2", "Bob", "bob@test.com")

	versionService := services.NewVersionService(fileRepo, versionRepo, userRepo)
	fileService := services.NewFileService(
		fileRepo, userRepo, storageProvider, versionRepo, searchEngine,
		userService, versionService,
	)
	folderService := services.NewFolderService(fileRepo, userRepo, searchEngine)
	sharingService := services.NewSharingService(fileRepo, permissionRepo)

	return sharingService, fileService, folderService
}

func TestSharingService_ShareFile(t *testing.T) {
	sharingService, fileService, folderService := setupSharingService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "shared.txt", rootFolder.ID, "text/plain", []byte("content"))

	err := sharingService.ShareFile("user1", file.ID, "user2", models.PermissionView)
	if err != nil {
		t.Fatalf("ShareFile failed: %v", err)
	}

	hasAccess, err := sharingService.HasAccess("user2", file.ID, false)
	if err != nil {
		t.Fatalf("HasAccess failed: %v", err)
	}
	if !hasAccess {
		t.Error("Expected user2 to have view access")
	}

	hasEdit, _ := sharingService.HasAccess("user2", file.ID, true)
	if hasEdit {
		t.Error("Expected user2 to NOT have edit access (View permission only)")
	}
}

func TestSharingService_ShareFile_EditPermission(t *testing.T) {
	sharingService, fileService, folderService := setupSharingService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "edit.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := sharingService.ShareFile("user1", file.ID, "user2", models.PermissionEdit)
	if err != nil {
		t.Fatalf("ShareFile failed: %v", err)
	}

	hasEdit, _ := sharingService.HasAccess("user2", file.ID, true)
	if !hasEdit {
		t.Error("Expected user2 to have edit access")
	}
}

func TestSharingService_ShareFile_OnlyOwnerCanShare(t *testing.T) {
	sharingService, fileService, folderService := setupSharingService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "owner.txt", rootFolder.ID, "text/plain", []byte("x"))

	err := sharingService.ShareFile("user2", file.ID, "user1", models.PermissionView)
	if err != services.ErrNotOwnerCannotShare {
		t.Errorf("Expected ErrNotOwnerCannotShare, got %v", err)
	}
}

func TestSharingService_ShareFolder(t *testing.T) {
	sharingService, _, folderService := setupSharingService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")

	err := sharingService.ShareFolder("user1", rootFolder.ID, "user2", models.PermissionView)
	if err != nil {
		t.Fatalf("ShareFolder failed: %v", err)
	}

	hasAccess, _ := sharingService.HasAccess("user2", rootFolder.ID, false)
	if !hasAccess {
		t.Error("Expected user2 to have access to shared folder")
	}
}

func TestSharingService_RevokeAccess(t *testing.T) {
	sharingService, fileService, folderService := setupSharingService(t)

	rootFolder, _ := folderService.CreateFolder("user1", "Root", "")
	file, _ := fileService.Upload("user1", "revoke.txt", rootFolder.ID, "text/plain", []byte("x"))
	_ = sharingService.ShareFile("user1", file.ID, "user2", models.PermissionView)

	err := sharingService.RevokeAccess("user1", file.ID, "user2")
	if err != nil {
		t.Fatalf("RevokeAccess failed: %v", err)
	}

	hasAccess, _ := sharingService.HasAccess("user2", file.ID, false)
	if hasAccess {
		t.Error("Expected user2 to NOT have access after revoke")
	}
}
