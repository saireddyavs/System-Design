# System Design: Dropbox / File Sync Service

## 1. Problem Statement & Requirements

### Problem Statement
Design a cloud-based file synchronization service that allows users to store files in the cloud and automatically sync them across multiple devices. Users can upload, download, share files, and collaborate with conflict resolution when multiple users edit the same file simultaneously.

### Functional Requirements
- **Upload/Download**: Users can upload files of any size and download them on any device
- **Automatic Sync**: Changes made on one device automatically propagate to all other devices
- **File Versioning**: Store and retrieve previous versions of files (e.g., last 30 days)
- **Sharing**: Share files/folders with other users (view-only or edit permissions)
- **Conflict Resolution**: Handle simultaneous edits from multiple users/devices
- **Search**: Search files by name and content (optional)
- **Offline Support**: Queue changes when offline, sync when reconnected

### Non-Functional Requirements
- **Scale**: 700M+ users, billions of files, petabytes of storage
- **Latency**: Sync notifications within seconds; upload/download optimized for large files
- **Availability**: 99.9% uptime for metadata; 99.99% for block storage
- **Consistency**: Strong consistency for metadata; eventual consistency acceptable for sync propagation
- **Security**: Encryption at rest and in transit; optional client-side encryption

### Security Requirements
- **Authentication**: JWT or OAuth; refresh token rotation
- **Authorization**: Per-file/share permissions; enforce on every access
- **Encryption**: TLS 1.3 in transit; AES-256 at rest (S3 SSE)
- **Audit**: Log access to shared files; detect anomalous download patterns

### Out of Scope
- Real-time collaborative editing (like Google Docs)
- Full-text search across file content (can be added later)
- Mobile-specific optimizations (battery, cellular data)
- Integration with third-party apps (OAuth, APIs)

---

## 2. Back-of-Envelope Estimation

### Capacity Planning Summary

Before diving into numbers, establish the key drivers: number of users, files per user, average file size, sync frequency. These determine storage, bandwidth, and compute needs.

### Assumptions
- 700M users, 20% active monthly = 140M MAU
- Average 1000 files per user = 140B files
- Average file size 2MB = 280 PB total storage
- Daily active users: 50M; each syncs 50 files/day = 2.5B sync operations/day

### QPS Estimates
| Operation | Daily Volume | QPS (peak 3x avg) |
|-----------|--------------|-------------------|
| File uploads | 2.5B | ~100,000 |
| Metadata reads | 10B | ~400,000 |
| Block storage reads | 5B | ~200,000 |
| Sync notifications | 2.5B | ~100,000 |

### Storage Estimates
- **Metadata DB**: 140B files × 1KB metadata ≈ 140 TB
- **Block Storage**: 280 PB (with deduplication, expect 40-60% savings → ~150 PB)
- **Version history**: 30 days × 5% churn × 280 PB ≈ 4 PB

### Bandwidth Estimates
- **Upload**: 2.5B × 2MB = 5 PB/day ≈ 60 GB/s peak
- **Download**: 3× upload (users access from multiple devices) ≈ 180 GB/s peak
- **With deduplication**: 60-70% bandwidth savings on uploads

### Cache Estimates
- **Metadata cache**: 20% hot files × 1KB × 140B = 28 TB (use Redis cluster)
- **Block cache**: CDN + edge cache for popular shared files
- **LRU eviction**: Keep most recently accessed chunks in memory

### Network Topology Considerations

- **Upload path**: Client → API Gateway → Block Service → S3. Bottleneck often at client or S3.
- **Download path**: Client → CDN (cache hit) or S3 (miss). CDN reduces origin load.
- **Sync path**: Client → API → Kafka → Sync workers. Low bandwidth; high message volume.

---

## 3. API Design

### REST Endpoints

```
# Authentication
POST   /api/v1/auth/login          # Returns JWT + refresh token
POST   /api/v1/auth/refresh        # Refresh JWT
POST   /api/v1/auth/logout         # Invalidate tokens

# File Metadata
GET    /api/v1/files/{path}        # Get file/folder metadata (recursive optional)
POST   /api/v1/files               # Create file/folder
PUT    /api/v1/files/{path}        # Update file metadata (rename, move)
DELETE /api/v1/files/{path}        # Delete file/folder
GET    /api/v1/files/{path}/versions  # List versions
GET    /api/v1/files/{path}/versions/{version_id}  # Get specific version

# Chunk Upload (Content-Addressable)
POST   /api/v1/blocks/check        # Batch check which chunks exist (send hashes)
POST   /api/v1/blocks/upload       # Upload chunk (hash in header, content in body)
GET    /api/v1/blocks/{hash}       # Download chunk by hash

# Sharing
POST   /api/v1/shares              # Create share link or invite user
GET    /api/v1/shares/{share_id}   # Get shared content
PUT    /api/v1/shares/{share_id}/permissions  # Update permissions
DELETE /api/v1/shares/{share_id}   # Revoke share

# Sync
GET    /api/v1/sync/delta         # Get changes since cursor (long polling)
POST   /api/v1/sync/commit        # Commit file with chunk list (atomic)
```

### WebSocket (Alternative to Long Polling)
```
WS /api/v1/sync/stream
- Client connects with auth token
- Server pushes: { type: "file_change", path: "/docs/report.pdf", version: 5 }
- Client fetches delta via REST after notification
```

### Key Request/Response Examples

**Check chunks (deduplication)**:
```json
POST /api/v1/blocks/check
{
  "hashes": ["sha256:abc123...", "sha256:def456...", "sha256:ghi789..."]
}
Response: {
  "existing": ["sha256:abc123..."],
  "missing": ["sha256:def456...", "sha256:ghi789..."]
}
```

**Commit file**:
```json
POST /api/v1/sync/commit
{
  "path": "/Documents/report.pdf",
  "parent_revision": "rev_12345",
  "chunks": [
    {"hash": "sha256:abc...", "size": 4194304},
    {"hash": "sha256:def...", "size": 4194304},
    {"hash": "sha256:ghi...", "size": 1048576}
  ],
  "size": 9437184,
  "mtime": "2024-03-10T14:30:00Z"
}
```

**Delta sync (long poll)**:
```json
GET /api/v1/sync/delta?cursor=cursor_xyz&timeout=30
Response (when changes exist or timeout):
{
  "cursor": "cursor_abc",
  "has_more": false,
  "entries": [
    {"path": "/docs/file.pdf", "type": "file", "deleted": false, "metadata": {...}}
  ]
}
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Metadata**: PostgreSQL or Cassandra (for scale) — relational for file tree, versions, sharing
- **Block index**: Cassandra or DynamoDB — hash → storage location mapping
- **Block storage**: Object storage (S3, GCS) — content-addressable by hash

### Schema

**users**
| Column | Type | Description |
|--------|------|-------------|
| user_id | UUID PK | User identifier |
| email | VARCHAR | Login email |
| storage_used | BIGINT | Bytes used |
| storage_limit | BIGINT | Quota |
| created_at | TIMESTAMP | |

**files** (metadata)
| Column | Type | Description |
|--------|------|-------------|
| file_id | UUID PK | Unique file ID |
| parent_id | UUID FK | Parent folder (null = root) |
| owner_id | UUID FK | User who owns |
| name | VARCHAR | File/folder name |
| path | VARCHAR | Full path (denormalized for queries) |
| is_folder | BOOLEAN | |
| size | BIGINT | File size in bytes |
| content_hash | VARCHAR | Hash of chunk list (for quick change detection) |
| version | INT | Incrementing version number |
| mtime | TIMESTAMP | Last modified |
| ctime | TIMESTAMP | Created |
| deleted | BOOLEAN | Soft delete |

**file_versions**
| Column | Type | Description |
|--------|------|-------------|
| version_id | UUID PK | |
| file_id | UUID FK | |
| version | INT | |
| chunk_list | JSONB | Ordered list of chunk hashes |
| size | BIGINT | |
| created_at | TIMESTAMP | |

**chunks** (block index)
| Column | Type | Description |
|--------|------|-------------|
| chunk_hash | VARCHAR PK | SHA-256 of chunk content |
| storage_key | VARCHAR | S3/GCS object key |
| size | INT | Chunk size |
| ref_count | INT | Deduplication reference count |

**shares**
| Column | Type | Description |
|--------|------|-------------|
| share_id | UUID PK | |
| file_id | UUID FK | |
| owner_id | UUID FK | |
| share_type | ENUM | link, user |
| sharee_id | UUID | User ID if user share |
| permission | ENUM | view, edit |
| expires_at | TIMESTAMP | Optional |

**sync_cursors**
| Column | Type | Description |
|--------|------|-------------|
| user_id | UUID PK | |
| device_id | VARCHAR PK | |
| cursor | VARCHAR | Last processed change ID |
| updated_at | TIMESTAMP | |

### Indexes

- `files(owner_id, path)` — List user's files
- `files(parent_id)` — List children of folder
- `files(path)` — Path lookup (if denormalized)
- `chunks(chunk_hash)` — Block lookup (primary key)
- `file_versions(file_id, version)` — Version retrieval
- `shares(file_id)`, `shares(sharee_id)` — Permission checks

### Partitioning Strategy

- **files**: Partition by `owner_id` hash for even distribution
- **chunks**: Partition by `chunk_hash` prefix (already sharded)
- **file_versions**: Partition by `file_id` for locality with files

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT (Desktop/Mobile/Web)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ File Watcher │  │ Chunk Splitter│  │ Hash Compute │  │ Sync Engine          │ │
│  │ (inotify)    │  │ (4MB chunks) │  │ (SHA-256)    │  │ (delta sync, merge)  │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┬─────────────┘ │
│         │                 │                 │                    │              │
└─────────┼─────────────────┼─────────────────┼────────────────────┼──────────────┘
          │                 │                 │                    │
          ▼                 ▼                 ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY / LOAD BALANCER                          │
└─────────────────────────────────────────────────────────────────────────────────┘
          │                 │                 │                    │
          ▼                 ▼                 ▼                    ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Metadata Service│ │ Block Service    │ │ Sync Service     │ │ Share Service    │
│ (file tree,     │ │ (chunk upload/   │ │ (delta,          │ │ (permissions,   │
│  versions,      │ │  download,       │ │  notifications)  │ │  links)          │
│  CRUD)          │ │  CAS by hash)    │ │                  │ │                  │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │                   │
         ▼                   ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ PostgreSQL/     │ │ Object Storage  │ │ Message Queue    │ │ Redis           │
│ Cassandra       │ │ (S3/GCS)        │ │ (Kafka)          │ │ (cache, pub/sub)│
│ (metadata)      │ │ Chunks by hash  │ │ Change events    │ │                 │
└─────────────────┘ └─────────────────┘ └─────────────────┘ └─────────────────┘

                              SYNC FLOW (ASCII)
┌──────────┐    1. Detect change     ┌──────────┐    2. Split, hash    ┌──────────┐
│ Client A │ ──────────────────────►│ Sync Eng │ ───────────────────►│ Chunks   │
│ (edit)   │                        │          │                      │ abc,def  │
└──────────┘                        └────┬─────┘                      └────┬─────┘
                                         │ 3. Check existing chunks         │
                                         ▼                                 │
                                  ┌──────────────┐    4. Upload missing    │
                                  │ Block Service│ ◄───────────────────────┘
                                  └──────┬───────┘
                                         │ 5. Commit (path, chunk list)
                                         ▼
                                  ┌──────────────┐    6. Publish event
                                  │Metadata Svc  │ ───────────────────────► Kafka
                                  └──────────────┘
                                         │
                                         │ 7. Notify Client B (WebSocket/poll)
                                         ▼
                                  ┌──────────────┐    8. Fetch delta, chunks
                                  │ Client B     │ ◄──────────────────────────
                                  └──────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Chunked File Upload & Deduplication

**Chunk Size**: 4MB (configurable). Trade-off: smaller = better deduplication, more metadata; larger = less overhead, worse deduplication.

**Process**:
1. Split file into 4MB chunks (last chunk may be smaller)
2. Compute SHA-256 hash of each chunk
3. Batch check which hashes already exist: `POST /blocks/check` with up to 1000 hashes
4. Upload only missing chunks to block storage
5. Commit file metadata with ordered list of chunk hashes

**Deduplication**: Same chunk (e.g., common header in PDFs) stored once. `ref_count` incremented when new file references it; decremented on delete. Chunk deleted when ref_count = 0.

**Compression**: Apply gzip/zstd to chunks before upload if size reduces. Store compressed size; decompress on download.

### 6.2 File Metadata Service

**Responsibilities**:
- CRUD for files/folders
- Maintain file tree (parent-child relationships)
- Version management (store chunk list per version)
- Path resolution and validation
- Enforce quotas (storage_used)

**Versioning**: Each commit creates a new version. Store `(file_id, version) → chunk_list`. Retrieve old version by fetching its chunk list and assembling chunks.

**Conflict Resolution**:
- **Last-writer-wins**: Use `parent_revision` in commit. If current version != parent_revision, reject and return conflict.
- **Conflict copy**: Create `report (conflicted copy).pdf` with user's version; notify user to merge manually.
- **Operational transform**: Out of scope for binary files; only for text (Google Docs style).

### 6.3 Block Storage Service (Content-Addressable Storage)

**Key**: `chunk_hash` (e.g., `sha256:abc123...`)
**Value**: Raw chunk bytes (optionally compressed)

**Benefits**:
- Deduplication: same content = same hash = single storage
- Integrity: hash verifies on download
- Immutable: chunks never updated, only referenced

**Sharding**: Shard by hash prefix (e.g., first 2 hex chars) for even distribution. Each shard maps to S3 bucket or prefix.

### 6.4 Sync Service

**Change Notification**:
- **Long polling**: Client calls `GET /sync/delta?cursor=X`. Server holds until new changes or timeout (30s). Returns list of changed files.
- **WebSocket**: Persistent connection; server pushes immediately on change. Lower latency, more connections to maintain.

**Delta Sync**: Client sends `cursor` (last processed change ID). Server returns all changes since cursor. Client merges into local state, fetches changed file metadata, downloads new chunks.

**Client-Side**:
- **File watcher**: inotify (Linux), FSEvents (Mac), ReadDirectoryChangesW (Windows)
- **Debouncing**: Batch rapid changes (e.g., 500ms) before uploading
- **Conflict detection**: Before commit, re-fetch current version. If changed, apply conflict resolution.

### 6.5 Encryption

**Server-side**: TLS in transit; AES-256 at rest (S3 SSE, GCS encryption).

**Client-side (optional)**:
- User has encryption key (or derived from password)
- Client encrypts chunks before upload; decrypts on download
- Server never sees plaintext; zero-knowledge architecture
- Trade-off: No server-side deduplication (different keys = different ciphertext for same plaintext)

### 6.6 Bandwidth Optimization

- **Delta sync**: Only transfer changed chunks
- **Deduplication**: Skip upload if chunk exists
- **Compression**: gzip/zstd for chunks
- **CDN**: Cache popular shared files at edge
- **Resumable uploads**: Chunk-level resume for large files (multipart upload)

### 6.7 Selective Sync (Smart Sync)

**Problem**: User has 1TB in cloud but only 256GB local disk. Cannot sync everything.

**Solution**:
- **Placeholder files**: Sync metadata only; file appears in folder but not downloaded
- **On-demand**: Download chunk when user opens file; evict when disk full (LRU)
- **Pinned**: User marks folder "always keep on this device" — full sync
- **Server**: Store `sync_scope` per user/device: `full`, `selective`, or path list
- **Delta**: Filter changes by sync scope before sending to client

### 6.8 Garbage Collection (Chunk Cleanup)

**Problem**: Deleted files leave orphan chunks; need to reclaim storage.

**Process**:
1. Track `ref_count` per chunk (incremented when file references, decremented when file deleted)
2. Background job: Find chunks with `ref_count = 0`
3. Delete from object storage; remove from chunk index
4. **Lazy deletion**: Delay 30 days before physical delete (recover from accidental delete)

**Challenges**: Race conditions when file deleted while upload in progress; use distributed lock.

### 6.9 Sharing Deep Dive

**Link sharing**:
- Generate share link with token (e.g., `https://dropbox.com/s/abc123/file.pdf`)
- Token maps to `(file_id, permission, expires_at)`
- No user account required for view; optional password protection
- **Bandwidth**: Shared files heavily accessed → CDN critical

**User sharing**:
- Add sharee to `shares` table with `view` or `edit` permission
- On access: Check if user in sharees or file in shared folder
- **Inheritance**: Folder share implies all children shared
- **Cache**: Per-user permission cache; invalidate on share change

### 6.10 Client Architecture (Desktop)

**Components**:
- **Index**: Local SQLite of file tree + chunk hashes (synced from server)
- **Watcher**: inotify/FSEvents → detect creates, modifies, deletes
- **Uploader**: Queue of pending uploads; batch chunk check; upload missing; commit
- **Downloader**: On demand or background; write to temp then atomic move
- **Conflict UI**: Show "conflicted copy" when LWW rejects; let user merge or keep

**Sync order**: Process server changes first (merge), then upload local changes. Prevents overwriting server with stale local.

---

## 7. Scaling

### Sharding

**Metadata**: Shard by `user_id` or `file_id`. User's files in same shard for locality. Use consistent hashing for rebalancing.

**Block index**: Shard by `chunk_hash` prefix. 256 shards (first 2 hex chars) = 256 partitions.

**Block storage**: S3/GCS handle scaling; use multiple buckets per region.

### Caching

- **Metadata cache**: Redis. Key: `user_id:path` or `file_id`. TTL: 5 min. Invalidate on write.
- **Chunk cache**: CDN for frequently accessed chunks (shared files). Edge cache with LRU.
- **Delta cache**: Cache recent deltas per user for fast replay.

### CDN

- Serve chunk downloads from edge locations
- Cache shared file chunks (high read ratio)
- Reduce latency for global users

### Read/Write Separation

- Metadata: Primary for writes; read replicas for reads (eventual consistency OK for listing)
- Block storage: Read-heavy; CDN absorbs most reads

### Multi-Region Considerations

**Metadata**: Active-active difficult (conflict resolution). Prefer active-passive with async replica in second region for DR.

**Block storage**: S3 Cross-Region Replication (CRR) for durability. Route reads to nearest region.

**Sync**: Deploy sync/API in multiple regions; route by user location. Shared Kafka cluster or per-region with replication.

### Throttling & Rate Limiting

- **Per-user upload**: Limit concurrent chunk uploads (e.g., 10) to prevent single user saturating
- **API rate limit**: 1000 req/min per user for metadata
- **Block upload**: 100 MB/s per user (configurable by plan)

---

## 8. Failure Handling

### Component Failures

| Component | Failure Mode | Mitigation |
|-----------|--------------|------------|
| Metadata DB | Primary down | Failover to replica (async replication, possible data loss window) |
| Block storage | S3 outage | Multi-region replication; serve from replica region |
| Sync service | Crash | Stateless; restart, clients reconnect |
| Message queue | Kafka down | Persist changes to DB; replay when Kafka recovers |

### Redundancy

- **Metadata**: Multi-AZ PostgreSQL; or Cassandra with RF=3
- **Block storage**: S3 cross-region replication; 11 nines durability
- **Sync**: Multiple sync service instances; stateless

### Client Resilience

- **Offline queue**: Store pending changes locally; retry on reconnect
- **Retry with backoff**: Exponential backoff for transient failures
- **Chunk upload**: Idempotent (same hash = overwrite); retry individual chunks

### Data Loss Prevention

- **WAL**: Database write-ahead log; replay on crash
- **Block storage**: Versioning enabled; recover from accidental delete
- **Backup**: Daily metadata backup; point-in-time recovery

### Split-Brain Scenarios

- **Metadata primary failover**: Promote replica; possible last few seconds of writes lost. Clients retry.
- **Block storage**: S3 multi-region; automatic failover. No split-brain (single namespace).
- **Kafka**: Consumer groups; at-least-once delivery. Idempotent processing handles duplicates.

### Recovery Procedures

1. **Metadata corruption**: Restore from backup; replay WAL; reconcile with block storage ref_counts
2. **Orphan chunks**: GC job finds ref_count=0; safe to delete after retention
3. **Client data loss**: User can restore from version history (last 30 days)

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| upload_latency_p99 | Time to upload file | > 5s |
| sync_notification_latency | Time from commit to client notified | > 10s |
| block_storage_error_rate | Failed chunk uploads/downloads | > 0.1% |
| metadata_db_latency | Query latency | > 100ms p99 |
| deduplication_ratio | (Logical size - physical size) / logical | Track trend |
| cache_hit_rate | Metadata cache hits | < 90% |

### Logging

- Structured logs: request_id, user_id, file_path, chunk_count, duration
- Audit log: file access, share creation, permission changes

### Tracing

- Distributed tracing (OpenTelemetry): Trace upload flow from client → API → metadata → block storage
- Trace sync: commit → Kafka → sync service → client notification

### Dashboards

- Real-time: QPS, error rate, latency by endpoint
- Capacity: Storage growth, chunk count, metadata size
- Business: DAU, files synced/day, storage per user

### SLOs and Error Budgets

- **Availability**: 99.9% = 43 min downtime/month. Reserve for deployments.
- **Latency**: p99 < 5s for upload; p99 < 2s for metadata read
- **Correctness**: Zero double-write; zero chunk loss (verified by hash)

### Alerting Runbooks

- **High upload latency**: Check block storage region; S3 throttling; client network
- **Sync delay**: Kafka lag; sync worker backlog; WebSocket connection drops
- **High error rate**: DB connection pool; Redis memory; API gateway limits

---

## 10. Interview Tips

### Follow-up Questions

1. **How do you handle a user uploading a 100GB file?** Chunked upload with resumable multipart; background job for very large files; consider streaming upload.
2. **How does sharing work with 10,000 users?** Share metadata stores list or use group; for link sharing, no limit. Enumerate permissions on access (cache per user).
3. **What if two users edit the same file simultaneously?** Last-writer-wins with version check; or create conflict copy. For text, operational transform.
4. **How do you implement "selective sync"?** Client sends list of paths to sync; server filters delta by path prefix. Lazy folder expansion.
5. **How do you handle renames?** Detect as delete + create (content hash match) to avoid re-upload. Or explicit rename API with move semantics.

### Common Mistakes

1. **Uploading entire file on every change**: Always use chunking and delta sync.
2. **No deduplication**: Content-addressable storage is key for scale.
3. **Synchronous sync**: Use async (queue, WebSocket) for notifications.
4. **Ignoring conflict resolution**: Must define strategy (LWW, conflict copy, OT).
5. **Single region**: Block storage and metadata need multi-region for durability and latency.

### What to Emphasize

- **Chunking + hashing**: Core to deduplication and bandwidth savings
- **Content-addressable storage**: Hash as key, immutable chunks
- **Delta sync**: Only transfer what changed
- **Separation of metadata and blocks**: Different scaling characteristics
- **Conflict resolution strategy**: Clear and explicit

### Sample Discussion Flow

1. **Clarify**: "Is this like Dropbox or Google Drive?" — Both; focus on sync, versioning, sharing.
2. **Start simple**: Single server, DB, object storage. Then scale out.
3. **Chunking first**: "How do we avoid re-uploading entire file?" → Chunk + hash + delta.
4. **Deduplication**: "Same file uploaded by 2 users?" → Same chunks, store once.
5. **Sync**: "How does Client B know file changed?" → Commit publishes event; WebSocket or poll.
6. **Conflict**: "Two edits at same time?" → Version check; LWW or conflict copy.

### Time-Boxed Approach (45 min interview)

- **0-5 min**: Clarify requirements, write down assumptions
- **5-15 min**: High-level design (client, API, metadata, block storage, sync)
- **15-25 min**: Deep dive chunking, deduplication, delta sync
- **25-35 min**: Scaling (sharding, caching), failure handling
- **35-45 min**: Q&A, follow-ups, trade-offs

### Additional Deep-Dive Topics

**Chunk size trade-offs**:
- 4MB: Good balance; 1GB file = 256 chunks; metadata overhead ~10KB
- 1MB: Better deduplication for small edits; 4× metadata
- 16MB: Less metadata; worse deduplication for documents

**Rename/move optimization**:
- Detect rename: Same content hash, different path. Update metadata only; no chunk transfer.
- Move folder: Update parent_id for all descendants; batch in single transaction.

**Quota enforcement**:
- On commit: `storage_used + new_size - old_size <= storage_limit`
- Lazy check: Allow slightly over; background job enforces; block upload when over.

**Bandwidth by region**:
- Users in Asia uploading to US datacenter: High latency. Deploy block storage in multiple regions; route by user.

### Design Alternatives Considered

**Full file upload vs chunked**: Full file simpler but no deduplication, no delta sync. Chunked wins at scale.

**Eventual vs strong consistency for metadata**: Strong for single-file ops (read your own file). Eventual OK for "list folder" across devices. We choose strong for correctness.

**Long poll vs WebSocket for sync**: Long poll simpler, works through firewalls. WebSocket lower latency, fewer connections. Support both; client picks.

**Centralized vs distributed metadata**: Centralized (PostgreSQL) simpler. Distributed (Cassandra) for extreme scale. Start centralized; shard when needed.

### Phased Rollout

**Phase 1 (MVP)**: Single-region upload/download, basic metadata, no versioning. Full file upload (no chunking). Prove out core flow.

**Phase 2**: Chunking + deduplication. Block storage. Delta sync. Versioning (last 7 days).

**Phase 3**: Sharing, conflict resolution, WebSocket. Long poll. Version history 30 days.

**Phase 4**: Multi-region, selective sync, client-side encryption. Scale to 100M users.

### Quick Reference Card

| Concept | Key Point |
|---------|-----------|
| Chunk size | 4MB; balance dedup vs metadata |
| Hash | SHA-256; content-addressable |
| Dedup | Same chunk = same hash = store once |
| Delta sync | Only changed chunks; parent_revision check |
| Conflict | LWW or conflict copy |
| Sync notify | WebSocket or long poll |
| Block storage | S3/GCS; key = hash |
| Metadata | PostgreSQL; file tree, versions, shares |

### Glossary

- **CAS**: Content-Addressable Storage — store by hash, not path
- **Chunk**: Fixed-size (4MB) block of file content
- **Delta sync**: Sync only changes since last cursor
- **LWW**: Last-Writer-Wins — conflict resolution strategy
- **Ref count**: Number of files referencing a chunk; for GC
