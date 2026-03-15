# File Storage System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm hierarchy (files/folders), upload/download, sharing levels, versioning, quota; scope out search, REST API |
| 2. Core Models | 7 min | FileSystemItem (interface), File, Folder, User, Permission, Version |
| 3. Repository Interfaces | 5 min | FileRepository, UserRepository, PermissionRepository, VersionRepository, StorageProvider |
| 4. Service Interfaces | 5 min | StorageProvider (Strategy), ShareObserver (Observer), SearchEngine |
| 5. Core Service Implementation | 12 min | FileService.Upload() — quota check, storage write, metadata, version; SharingService.ShareFile() |
| 6. main.go Wiring | 5 min | Repos, LocalStorageProvider, services with DI, ShareObserver registration |
| 7. Extend & Discuss | 8 min | Composite pattern (Folder.GetSize), Strategy for S3, Observer for share notifications |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- File/folder hierarchy? → Tree structure (folders contain files and sub-folders)
- Storage backend? → Local for LLD; S3/GCS swappable via Strategy
- Sharing model? → View, Edit, Owner permission levels; only owner can share/delete
- Versioning? → Keep last N versions per file (e.g., 10)
- Storage quota per user? → Default 1GB; track UsedStorage
- Search required? → Simple name search; full-text out of scope

**Scope out:** Full-text search (Elasticsearch), real-time collaboration, encryption at rest, REST API layer.

## Phase 2: Core Models (7 min)

**Start with (Composite pattern):**
- `FileSystemItem` interface: `GetID()`, `GetName()`, `GetOwnerID()`, `GetParentFolderID()`, `GetSize()`, `IsFolder()`
- `File`: implements FileSystemItem; ID, Name, OwnerID, ParentFolderID, Size, MimeType, Content, Versions
- `Folder`: implements FileSystemItem; Children `[]FileSystemItem`; `GetSize()` recursively sums children

**Then:**
- `User`: ID, Name, Email, StorageQuota, UsedStorage
- `Permission`: FileID, UserID, Level (View/Edit/Owner), GrantedBy
- `Version`: FileID, VersionNumber, Content, Size

**Skip for now:** SearchEngine implementation, Observer details, full permission inheritance (parent folder).

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `FileRepository`: CreateFile, GetFileByID, UpdateFile, DeleteFile; CreateFolder, GetFolderByID, UpdateFolder, DeleteFolder; UpdateFileParent, UpdateFolderParent
- `UserRepository`: Create, GetByID, Update
- `PermissionRepository`: Create, GetByFileAndUser, DeleteByFileAndUser
- `VersionRepository`: Create, GetByFileID, DeleteByFileID
- `StorageProvider` (Strategy): Upload(path, content), Download(path), Delete(path)

**Key abstraction:** StorageProvider lets you swap LocalStorage for S3 without touching FileService.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `StorageProvider`: Upload, Download, Delete — Strategy for storage backend
- `ShareObserver`: OnFileShared(file, permission), OnFolderShared(folder, permission) — Observer for notifications
- `SearchEngine`: Index(item, path), SearchByName(query), RemoveFromIndex(fileID) — optional for LLD

**Key abstraction:** ShareObserver decouples sharing from email/audit; add observers without changing SharingService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `FileService.Upload(userID, name, parentFolderID, mimeType, content)` — this is where the core logic lives.

**Flow:**
1. Check quota: `userService.CheckQuota(userID, len(content))` — fail if UsedStorage + size > Quota
2. Validate parent folder exists (if parentFolderID != "")
3. Create File model, generate ID
4. `storageProvider.Upload(userID+"/"+fileID, content)` — store blob
5. `fileRepo.CreateFile(file)` — persist metadata; rollback storage on failure
6. Update user: `user.AddUsedStorage(size)`
7. `searchEngine.Index(file, path)` — for search (optional)
8. (Versioning: on update, VersionService.CreateVersion before overwrite; cap at 10)

**SharingService.ShareFile(ownerID, fileID, targetUserID, level):**
1. Get file, verify ownerID == file.OwnerID
2. Check not already shared
3. Create Permission, persist
4. Notify observers: `obs.OnFileShared(file, permission)`

**Concurrency:** Use `sync.RWMutex` on FileService and repositories; Lock for write, RLock for read.

## Phase 6: main.go Wiring (5 min)

```go
userRepo := repositories.NewInMemoryUserRepo()
fileRepo := repositories.NewInMemoryFileRepo()
permissionRepo := repositories.NewInMemoryPermissionRepo()
versionRepo := repositories.NewInMemoryVersionRepo()
searchEngine := repositories.NewInMemorySearchEngine()
storageProvider := storage.NewLocalStorageProvider("")

userService := services.NewUserService(userRepo)
versionService := services.NewVersionService(fileRepo, versionRepo, userRepo)
fileService := services.NewFileService(fileRepo, userRepo, storageProvider, versionRepo, searchEngine, userService, versionService)
folderService := services.NewFolderService(fileRepo, userRepo, searchEngine)
sharingService := services.NewSharingService(fileRepo, permissionRepo)
sharingService.RegisterObserver(&ShareNotificationObserver{})
```

Show: StorageProvider (Strategy) injected; ShareObserver registered; all services depend on interfaces.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Composite**: File and Folder implement FileSystemItem; Folder.GetSize() recursively sums children; uniform treatment for list/delete
- **Strategy**: StorageProvider — swap Local/S3 without changing FileService
- **Observer**: ShareObserver — add email, audit log without modifying SharingService
- **Repository**: Abstract data access; swap in-memory for PostgreSQL

**Extensions:**
- S3StorageProvider implementation
- Permission inheritance (folder share → children)
- Version diff/merge
- Optimistic locking (ETags) for concurrent edits

## Tips

- **Prioritize if low on time:** FileSystemItem interface + Folder.GetSize() (Composite), Upload flow with quota, ShareFile with Permission levels. Skip SearchEngine, full versioning.
- **Common mistakes:** Forgetting quota check before upload; not rolling back storage on CreateFile failure; treating File and Folder differently (Composite unifies them).
- **What impresses:** Explaining *why* Composite (recursive size, uniform ops); StorageProvider Strategy for backend swap; ShareObserver for decoupled notifications; RWMutex for thread safety.
