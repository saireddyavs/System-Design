package main

import (
	"fmt"
	"log"

	"file-storage-system/internal/models"
	"file-storage-system/internal/repositories"
	"file-storage-system/internal/services"
	"file-storage-system/internal/storage"
)

func main() {
	// Initialize repositories (Dependency Inversion - depend on interfaces)
	userRepo := repositories.NewInMemoryUserRepo()
	fileRepo := repositories.NewInMemoryFileRepo()
	permissionRepo := repositories.NewInMemoryPermissionRepo()
	versionRepo := repositories.NewInMemoryVersionRepo()
	searchEngine := repositories.NewInMemorySearchEngine()

	// Storage provider (Strategy pattern - can swap for S3, GCS)
	storageProvider := storage.NewLocalStorageProvider("")

	// Services
	userService := services.NewUserService(userRepo)
	versionService := services.NewVersionService(fileRepo, versionRepo, userRepo)
	fileService := services.NewFileService(
		fileRepo, userRepo, storageProvider, versionRepo, searchEngine,
		userService, versionService,
	)
	folderService := services.NewFolderService(fileRepo, userRepo, searchEngine)
	sharingService := services.NewSharingService(fileRepo, permissionRepo)
	searchService := services.NewSearchService(searchEngine)

	// Register share observer (Observer pattern)
	sharingService.RegisterObserver(&ShareNotificationObserver{})

	// Demo: Create user
	user, err := userService.CreateUser("user1", "Alice", "alice@example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created user: %s (%s)\n", user.Name, user.Email)
	fmt.Printf("Storage quota: %d MB\n", user.StorageQuota/(1024*1024))

	// Demo: Create root folder
	rootFolder, err := folderService.CreateFolder("user1", "My Drive", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created folder: %s (ID: %s)\n", rootFolder.Name, rootFolder.ID)

	// Demo: Upload file
	content := []byte("Hello, File Storage System!")
	file, err := fileService.Upload("user1", "hello.txt", rootFolder.ID, "text/plain", content)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Uploaded file: %s (Size: %d bytes)\n", file.Name, file.Size)

	// Demo: Share file
	err = sharingService.ShareFile("user1", file.ID, "user2", models.PermissionView)
	if err != nil {
		fmt.Printf("Share (user2 may not exist): %v\n", err)
	} else {
		fmt.Println("Shared file with user2 (View permission)")
	}

	// Demo: Search
	results, _ := searchService.SearchByName("user1", "hello")
	fmt.Printf("Search 'hello': found %d result(s)\n", len(results))
	for _, r := range results {
		fmt.Printf("  - %s\n", r.Item.GetName())
	}

	// Demo: Get version history
	versions, _ := versionService.GetVersionHistory(file.ID)
	fmt.Printf("Version history: %d version(s)\n", len(versions))

	fmt.Println("\nFile Storage System demo completed successfully!")
}

// ShareNotificationObserver implements ShareObserver for demo.
type ShareNotificationObserver struct{}

func (o *ShareNotificationObserver) OnFileShared(file *models.File, permission *models.Permission) {
	fmt.Printf("[Observer] File '%s' shared with user %s (%s)\n", file.Name, permission.UserID, permission.Level)
}

func (o *ShareNotificationObserver) OnFolderShared(folder *models.Folder, permission *models.Permission) {
	fmt.Printf("[Observer] Folder '%s' shared with user %s (%s)\n", folder.Name, permission.UserID, permission.Level)
}
