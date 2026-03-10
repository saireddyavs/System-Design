# File Storage System (Google Drive-like) - Low Level Design

A production-quality, interview-ready LLD implementation of a file storage system in Go, following clean architecture and SOLID principles.

## 1. Problem Description

Design a file storage system similar to Google Drive with:
- **User management** with storage quotas
- **File and folder hierarchy** (tree structure)
- **File operations**: upload, download, delete, move, rename
- **File sharing** with permission levels (View, Edit, Owner)
- **File versioning** (keep history of changes)
- **Search** files by name
- **Storage quota tracking** per user

## 2. Requirements

| Requirement | Implementation |
|-------------|----------------|
| User management | UserService + UserRepository |
| Storage quotas | Default 1GB, tracked in User.UsedStorage |
| File/folder hierarchy | Composite pattern (FileSystemItem interface) |
| Upload/Download/Delete | FileService + StorageProvider |
| Move/Rename | FileService, FolderService |
| Sharing | SharingService + PermissionRepository |
| Permissions | View, Edit, Owner levels |
| Versioning | VersionService, keep last 10 versions |
| Search | SearchEngine interface + InMemorySearchEngine |
| Thread safety | sync.RWMutex on all repositories and services |

## 3. Core Entities & Relationships

```
┌─────────┐       owns       ┌──────────┐
│  User   │─────────────────▶│   File   │
│         │                   │          │
│ - ID    │       owns       │ - ID     │
│ - Name  │─────────────────▶│ - Name   │
│ - Email │                   │ - OwnerID│
│ - Quota │                   │ - Size  │
│ - Used  │                   │ - Content│
└─────────┘                   └────┬─────┘
       │                           │
       │ has                       │ has
       ▼                           ▼
┌─────────────┐             ┌──────────┐
│ Permission  │             │ Version  │
│ - FileID    │             │ - FileID │
│ - UserID    │             │ - Content│
│ - Level     │             │ - Number │
└─────────────┘             └──────────┘

┌─────────┐    parent     ┌────────┐
│  User   │─────────────▶│ Folder │
│         │              │        │
└─────────┘              │ - ID   │
                          │ - Name │
                          │ - Children (FileSystemItem[])
                          └────────┘
```

### Entity Details

| Entity | Key Fields | Purpose |
|--------|------------|---------|
| **User** | ID, Name, Email, StorageQuota, UsedStorage | Identity + quota tracking |
| **File** | ID, Name, OwnerID, ParentFolderID, Size, MimeType, Content | File metadata + content |
| **Folder** | ID, Name, OwnerID, ParentFolderID, Children | Container (Composite) |
| **Permission** | FileID, UserID, Level, GrantedBy | Sharing access control |
| **Version** | FileID, VersionNumber, Content, Size | Version history |

## 4. Composite Pattern - File System Hierarchy

**Why Composite?** Files and folders form a tree. Both need to support:
- Recursive size calculation (folder = sum of children)
- Recursive deletion
- Permission inheritance
- Uniform treatment in search/listing

**Implementation:**

```go
type FileSystemItem interface {
    GetID() string
    GetName() string
    GetOwnerID() string
    GetParentFolderID() string
    GetSize() int64      // File: returns Size; Folder: recursively sums children
    GetCreatedAt() time.Time
    IsFolder() bool
}

// File implements FileSystemItem
// Folder implements FileSystemItem, contains []FileSystemItem (Children)
```

**Benefits:**
- Single interface for files and folders
- Folder.GetSize() recursively calls child.GetSize()
- Easy to add new operations (e.g., GetPath, ListRecursive)
- Client code treats both uniformly

## 5. Design Patterns with WHY

| Pattern | Where | Why |
|---------|-------|-----|
| **Composite** | File + Folder → FileSystemItem | Tree structure; recursive ops (size, delete) |
| **Strategy** | StorageProvider (Local, S3, GCS) | Swap storage backend without changing business logic |
| **Observer** | ShareObserver on SharingService | Decouple sharing from notifications (email, audit log) |
| **Factory** | NewFile, NewFolder | Encapsulate creation; ensure valid state |
| **Repository** | FileRepository, UserRepository | Abstract data access; testable; swappable (DB vs in-memory) |

### Strategy - StorageProvider

```go
type StorageProvider interface {
    Upload(path string, content []byte) (string, error)
    Download(path string) ([]byte, error)
    Delete(path string) error
}
// Implementations: LocalStorageProvider, S3StorageProvider (future)
```

### Observer - Share Notifications

```go
type ShareObserver interface {
    OnFileShared(file *File, permission *Permission)
    OnFolderShared(folder *Folder, permission *Permission)
}
// SharingService.RegisterObserver(observer)
// On share: for _, obs := range observers { obs.OnFileShared(...) }
```

## 6. SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | FileService (file ops), FolderService (folder ops), SharingService (sharing), VersionService (versions) |
| **O - Open/Closed** | StorageProvider: add S3 without modifying FileService; ShareObserver: add new observers without changing SharingService |
| **L - Liskov Substitution** | File and Folder both substitute for FileSystemItem; any StorageProvider works |
| **I - Interface Segregation** | Separate interfaces: FileRepository, UserRepository, StorageProvider, SearchEngine (clients depend only on what they need) |
| **D - Dependency Inversion** | Services depend on interfaces (UserRepository), not concrete InMemoryUserRepo |

## 7. Permission Model

| Level | View | Edit | Delete | Share |
|-------|------|------|--------|-------|
| **View** | ✓ | ✗ | ✗ | ✗ |
| **Edit** | ✓ | ✓ | ✗ | ✗ |
| **Owner** | ✓ | ✓ | ✓ | ✓ |

- **Inheritance**: Sub-folders inherit parent permissions (simplified in LLD; full impl would walk parent chain)
- **Only owner** can share or delete
- **Shared users** get View or Edit based on permission level

## 8. Business Rules

- **Storage quota**: Default 1GB per user
- **File versioning**: Keep last 10 versions per file
- **Permission inheritance**: Sub-folders inherit parent permissions
- **Only owner** can share or delete
- **Shared users** can view or edit based on permission level

## 9. Interview Explanations

### 3-Minute Summary

"We built a Google Drive-like file storage system in Go. The core design uses the **Composite pattern** so files and folders both implement `FileSystemItem`, enabling recursive size calculation and uniform handling. We use **Repository pattern** for data access, **Strategy** for storage backends (local/S3), and **Observer** for share notifications. Services are split by responsibility (File, Folder, Sharing, Version), and we depend on interfaces for testability. Storage quotas are enforced on upload, and we keep the last 10 versions per file. Permissions support View, Edit, and Owner levels."

### 10-Minute Deep Dive

1. **Composite**: File and Folder implement FileSystemItem. Folder contains `[]FileSystemItem` (children). `GetSize()` on a folder recursively sums children. This allows listing, searching, and deleting subtrees uniformly.

2. **Clean Architecture**: Models in `models/`, interfaces in `interfaces/`, implementations in `repositories/` and `storage/`. Services orchestrate; they don't know about DB or storage details.

3. **SOLID**: Each service has one job. We add new storage via StorageProvider without touching FileService. ShareObserver lets us add email/audit without changing SharingService.

4. **Thread Safety**: All repos and services use `sync.RWMutex` for concurrent access.

5. **Quota**: User has StorageQuota and UsedStorage. On upload we check `GetAvailableStorage() >= size`; on delete we decrement UsedStorage.

6. **Versioning**: VersionRepository stores versions per file. CreateVersion is called on update. We cap at 10 versions (oldest evicted).

7. **Sharing**: Permission has FileID, UserID, Level. HasAccess checks owner or permission. Only owner can Share/Revoke.

## 10. Future Improvements

| Area | Improvement |
|------|-------------|
| **Storage** | S3/GCS StorageProvider; chunked upload for large files |
| **Search** | Elasticsearch/Meilisearch; full-text, metadata filters |
| **Permissions** | Inherited permissions; group-based sharing |
| **Versioning** | Delta storage; version diff/merge |
| **Concurrency** | Optimistic locking; ETags for conflict detection |
| **Persistence** | PostgreSQL + file blob storage |
| **API** | REST/gRPC; pagination, streaming |
| **Security** | Encryption at rest; signed URLs for download |

## 11. Running the Project

```bash
# Build
go build ./...

# Run demo
go run ./cmd/main.go

# Run tests
go test ./tests/... -v
```

## 12. Directory Structure

```
08-file-storage-system/
├── cmd/main.go                 # Demo entry point
├── internal/
│   ├── models/                 # User, File, Folder, Permission, Version
│   ├── interfaces/             # Repository, Storage, Search contracts
│   ├── services/              # Business logic
│   ├── repositories/          # In-memory implementations
│   ├── storage/               # LocalStorageProvider
│   └── utils/                 # ID generation
├── tests/                     # Unit tests
├── go.mod
└── README.md
```
