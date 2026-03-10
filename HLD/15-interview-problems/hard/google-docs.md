# Design Google Docs / Collaborative Editing

## 1. Problem Statement & Requirements

### Problem Statement
Design a real-time collaborative document editing system where multiple users can simultaneously edit the same document, see each other's cursors and selections, add comments, and maintain version history with undo/redo support.

### Functional Requirements
- **Real-time editing**: Multiple users edit simultaneously; changes appear in < 100 ms
- **Document model**: Rich text (bold, italic, headings, lists)

- **Conflict resolution**: Concurrent edits merge correctly without data loss
- **Presence**: Show who's viewing/editing; cursor and selection positions
- **Comments**: Inline comments, replies, resolve
- **Version history**: View and restore previous versions
- **Sharing**: Owner, editor, viewer, commenter permissions
- **Auto-save**: Persist changes automatically
- **Offline editing**: Edit offline; sync when back online

### Non-Functional Requirements
- **Latency**: < 100 ms for operation propagation
- **Availability**: 99.99%
- **Consistency**: Strong consistency for document state; eventual for presence
- **Scale**: Millions of concurrent documents; 10–50 editors per document

### Out of Scope
- Real-time collaboration on spreadsheets (complex cell dependencies)
- Advanced formatting (tables, images — simplified)
- Full offline conflict resolution (basic sync only)

---

## 2. Back-of-Envelope Estimation

### Assumptions
- 100M active users
- 500M documents
- 10M active editing sessions
- 5 editors per active document on average
- 1 operation per second per editor

### QPS Estimates
| Component | Calculation | QPS |
|-----------|-------------|-----|
| Operation broadcast | 10M × 5 × 1 | 50M |
| Document load | 10M / 3600 | ~2,800 |
| Presence updates | 10M × 5 × 0.5 | 25M |
| Version save | 10M × 0.1 | 1M |

### Storage (1 year)
| Data | Size | Count | Total |
|------|------|-------|-------|
| Document content | 50 KB avg | 500M | 25 TB |
| Version snapshots | 50 KB | 10B | 500 TB |
| Operations log | 200 B | 1.5T | 300 TB |
| Comments | 500 B | 5B | 2.5 TB |

### Bandwidth
- Operations: 50M × 500 B ≈ 25 GB/s (internal)
- Presence: 25M × 100 B ≈ 2.5 GB/s

### Cache
- Active document buffer: 10M × 50 KB = 500 GB
- Operation history (last 1000 ops): 10M × 1000 × 200 B = 2 TB

---

## 3. API Design

### REST Endpoints

```
GET    /api/v1/documents/{docId}           # Get document (with content)
POST   /api/v1/documents                   # Create document
PATCH  /api/v1/documents/{docId}           # Update metadata
DELETE /api/v1/documents/{docId}           # Delete document

GET    /api/v1/documents/{docId}/versions  # List versions
GET    /api/v1/documents/{docId}/versions/{v}  # Get version content
POST   /api/v1/documents/{docId}/restore  # Restore to version

GET    /api/v1/documents/{docId}/comments # List comments
POST   /api/v1/documents/{docId}/comments # Add comment
PATCH  /api/v1/documents/{docId}/comments/{cId}  # Resolve/reply

GET    /api/v1/documents/{docId}/permissions  # List permissions
POST   /api/v1/documents/{docId}/share     # Share with user/group
POST   /api/v1/documents/{docId}/unshare   # Revoke access
```

### WebSocket Endpoints

```
WS /ws/documents/{docId}   # Operation sync, presence, cursor updates
```

### WebSocket Message Types

```json
// Operation (client → server)
{"type": "op", "op": {"id": "op_1", "pos": 5, "opType": "insert", "content": "hello"}}

// Operation (server → clients)
{"type": "op", "op": {...}, "clientId": "user_123"}

// Presence (client → server)
{"type": "presence", "cursor": {"pos": 10}, "selection": {"start": 5, "end": 15}}

// Presence (server → clients)
{"type": "presence", "userId": "u_1", "cursor": {...}, "selection": {...}}

// Cursor (client → server)
{"type": "cursor", "pos": 20}
```

---

## 4. Data Model / Database Schema

### Document Model

**Rich text as tree structure** (e.g., ProseMirror-style):

```
Document
├── Paragraph
│   ├── Text("Hello ")
│   ├── Strong(Text("world"))
│   └── Text("!")
├── Heading(level=1)
│   └── Text("Title")
└── List
    ├── ListItem(Text("Item 1"))
    └── ListItem(Text("Item 2"))
```

**Operation representation** (OT-style):
- `Insert(pos, content)` — insert at position
- `Delete(pos, length)` — delete range
- `Retain(length)` — skip (for alignment)

### Database Tables

**documents**
| Column | Type |
|--------|------|
| id | UUID PK |
| title | VARCHAR |
| owner_id | UUID FK |
| content_snapshot | JSONB/TEXT |
| version | INT |
| created_at | TIMESTAMP |
| updated_at | TIMESTAMP |

**document_operations**
| Column | Type |
|--------|------|
| doc_id | UUID PK |
| seq | BIGINT PK |
| client_id | UUID |
| op_json | JSONB |
| applied_at | TIMESTAMP |

**document_versions**
| Column | Type |
|--------|------|
| doc_id | UUID PK |
| version | INT PK |
| content_snapshot | JSONB |
| created_at | TIMESTAMP |

**document_permissions**
| Column | Type |
|--------|------|
| doc_id | UUID PK |
| user_id | UUID PK |
| role | ENUM (owner, editor, viewer, commenter) |

**comments**
| Column | Type |
|--------|------|
| id | UUID PK |
| doc_id | UUID FK |
| anchor_start | INT |
| anchor_end | INT |
| author_id | UUID |
| content | TEXT |
| resolved | BOOLEAN |
| parent_id | UUID FK |
| created_at | TIMESTAMP |

### DB Choice
- **PostgreSQL**: Documents, versions, permissions, comments; ACID
- **Redis**: Active document buffer, presence, recent operations
- **Cassandra** (optional): Operation log for high write throughput

---

## 5. High-Level Architecture

```
                              ┌─────────────────────────────────────────┐
                              │            LOAD BALANCER                 │
                              └─────────────────────┬───────────────────┘
                                                    │
                    ┌───────────────────────────────┼───────────────────────────────┐
                    │                               │                               │
                    ▼                               ▼                               ▼
         ┌──────────────────┐            ┌──────────────────┐            ┌──────────────────┐
         │   REST API       │            │  WebSocket Server │            │  WebSocket Server│
         │   (CRUD, Share)  │            │  (Doc Sync)       │            │  (Doc Sync)      │
         └────────┬─────────┘            └────────┬─────────┘            └────────┬─────────┘
                  │                                │                                │
                  └────────────────────────────────┼────────────────────────────────┘
                                                   │
         ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
         │                                         │                                         │
         ▼                                         ▼                                         ▼
┌─────────────────┐                      ┌─────────────────┐                      ┌─────────────────┐
│ Document Service│                      │ Collaboration   │                      │ Presence Service│
│ (CRUD, Versions)│                      │ Service (OT/    │                      │ (Cursors, Who's │
│                 │                      │ CRDT)           │                      │  viewing)       │
└────────┬────────┘                      └────────┬────────┘                      └────────┬────────┘
         │                                         │                                         │
         │                                         ▼                                         │
         │                                ┌─────────────────┐                                │
         │                                │ Operation Log   │                                │
         │                                │ (Kafka/Redis)   │                                │
         │                                └────────┬────────┘                                │
         │                                         │                                         │
         └─────────────────────────────────────────┼─────────────────────────────────────────┘
                                                   │
         ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
         │                                         │                                         │
         ▼                                         ▼                                         ▼
┌─────────────────┐                      ┌─────────────────┐                      ┌─────────────────┐
│  PostgreSQL     │                      │  Redis          │                      │  Redis Pub/Sub  │
│  (Documents,    │                      │  (Doc buffer,   │                      │  (Presence)     │
│   Versions)     │                      │   recent ops)   │                      │                 │
└─────────────────┘                      └─────────────────┘                      └─────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Real-Time Collaboration: OT vs CRDT

#### Operational Transformation (OT)

**Idea**: When two operations are concurrent, *transform* one against the other so both can be applied without conflict.

**Transformation rules** (simplified for plain text):

```
T(Insert(p1, c1), Insert(p2, c2)):
  if p1 < p2: return Insert(p2 + len(c1), c2)
  if p1 > p2: return Insert(p2, c2)
  else: tie-break by client_id

T(Insert(p1, c1), Delete(p2, len)):
  if p1 <= p2: return Delete(p2 + len(c1), len)
  if p1 >= p2 + len: return Delete(p2, len)
  else: split Delete (complex)

T(Delete(p1, len1), Insert(p2, c2)): ...similar
T(Delete(p1, len1), Delete(p2, len2)): ...similar
```

**OT Flow**:
1. Client A sends Op1 (insert "X" at 5)
2. Client B sends Op2 (insert "Y" at 3) before receiving Op1
3. Server receives Op1, applies to doc
4. Server receives Op2; must transform against Op1: Op2' = T(Op2, Op1)
5. Op2' becomes "insert Y at 6" (because Op1 inserted 1 char at 5)
6. Apply Op2'; broadcast both Op1, Op2' to all clients

**OT Diagram**:

```
Client A                    Server                     Client B
   |                          |                           |
   |-- Op1: Ins(5,"X") ------>|                           |
   |                          |-- apply Op1               |
   |                          |<-- Op2: Ins(3,"Y") ------|
   |                          |-- Op2' = T(Op2, Op1)     |
   |                          |-- apply Op2'             |
   |<-- Op2' (transformed) ---|------------------------->| Op1
   |-- apply Op2'             |                          |-- apply Op1
   |                          |                          |
   Both clients converge to same state: "abcXYdef" (assuming doc was "abcdef")
```

**Pros**: Centralized, well-understood, used by Google Docs  
**Cons**: Complex transformation rules; hard to extend to rich text

---

#### Conflict-Free Replicated Data Types (CRDT)

**Idea**: Structure data so concurrent updates can be merged without transformation; merge is commutative and associative.

**Example: RGA (Replicated Growable Array)** for text:
- Each character has unique ID: (site_id, seq_num)
- Insert: add (id, char, after_id)
- Delete: tombstone (mark deleted, don't remove)
- Merge: union of all inserts; apply tombstones; order by (after_id, id)

**CRDT Merge Diagram**:

```
Site A (user 1):          Site B (user 2):
Doc: "Hi"                 Doc: "Hi"
Op: Ins("!", after "i")   Op: Ins("?", after "H")

After merge (both sites):
- A has: H(1,1), i(1,2), !(1,3)
- B has: H(2,1), ?(2,2), i(2,3)

Merge = union:
  H(1,1), H(2,1) — same char, one id
  i(1,2), i(2,3)
  ?(2,2) after H
  !(1,3) after i

Ordering: H, ?, i, !  (or H, i, ?, ! depending on tie-break)
Result: "Hi!?" or "Hi?!"
```

**CRDT Flow**:
1. No central server for merge; each client has full CRDT state
2. On edit: generate CRDT op (insert/delete with IDs)
3. Broadcast op to other clients (via server)
4. Each client merges op into local state (deterministic)
5. No transformation; merge is always defined

**OT vs CRDT Comparison**:

| Aspect | OT | CRDT |
|--------|-----|------|
| **Conflict resolution** | Transform ops before apply | Merge is always defined |
| **Central server** | Required (authoritative) | Optional (peer-to-peer possible) |
| **Operation size** | Small (pos, content) | Larger (IDs, metadata) |
| **Memory** | O(doc size) | O(doc size + tombstones) |
| **Undo** | Complex (inverse ops) | Easier (remove/undo op) |
| **Offline** | Hard (need server) | Natural (merge when online) |
| **Used by** | Google Docs, Etherpad | Figma, Automerge, Yjs |

### 6.2 Document Model (Tree Structure)

For rich text, use a tree (e.g., ProseMirror schema):

```json
{
  "type": "doc",
  "content": [
    {"type": "paragraph", "content": [{"type": "text", "text": "Hello"}]},
    {"type": "heading", "attrs": {"level": 1}, "content": [{"type": "text", "text": "Title"}]}
  ]
}
```

Positions: path in tree, e.g., `[0, 2]` = paragraph 0, offset 2.

### 6.3 Cursor and Selection Tracking

- Each client sends `{cursor: pos, selection: {start, end}}` on change
- Server broadcasts to other clients (exclude sender)
- Throttle: max 10 updates/sec per user
- Store in Redis: `presence:{docId}` → map of userId → cursor/selection
- TTL: 30 seconds (heartbeat)

### 6.4 Presence (Who's Viewing)

- On document open: join "presence" channel for docId
- Send heartbeat every 5 seconds
- Server maintains set of active users per document
- Broadcast join/leave to all viewers
- Show avatars, names in UI

### 6.5 Version History

**Snapshot approach**:
- Every N operations (e.g., 100) or every M minutes: save full document snapshot
- Store in `document_versions` with version number
- On "View history": load snapshot for selected version

**Diff approach** (optional):
- Store operations; replay to reconstruct any version
- More storage-efficient; slower to reconstruct

**Undo/Redo**:
- OT: Maintain inverse operations; transform against new ops
- CRDT: Track "undo" as new op that reverses effect

### 6.6 Conflict Resolution

- **OT**: Server transforms all incoming ops; single linear history
- **CRDT**: Merge is deterministic; no explicit conflict resolution
- **Last-write-wins** (avoid for text): Would lose data

### 6.7 Offline Editing and Sync

- **CRDT**: Ideal — buffer ops locally; merge when online
- **OT**: Queue ops; send to server when online; server transforms against missed ops; may require full doc fetch if offline too long

### 6.8 Permissions

| Role | View | Edit | Comment | Share |
|------|------|------|---------|-------|
| Owner | ✓ | ✓ | ✓ | ✓ |
| Editor | ✓ | ✓ | ✓ | ✗ |
| Commenter | ✓ | ✗ | ✓ | ✗ |
| Viewer | ✓ | ✗ | ✗ | ✗ |

- Check permission on document load and on each write
- Cache permissions in session

### 6.9 Auto-Save

- Debounce: 2 seconds after last edit
- Or: every 50 operations
- Write snapshot to PostgreSQL
- Async; don't block editing

---

## 7. Scaling

### Sharding
- **Documents**: By doc_id (UUID)
- **Operations**: By doc_id
- **Presence**: By doc_id (Redis key)

### Caching
- **Redis**: Active document content (last 1000 ops applied)
- **Redis**: Presence per document
- **CDN**: Static assets

### WebSocket Scaling
- Sticky sessions (doc_id → server instance)
- Redis Pub/Sub: Publish to channel `doc:{docId}`; all servers subscribed forward to connected clients

### Read Replicas
- PostgreSQL replicas for document load
- Primary for writes

---

## 8. Failure Handling

### Server Crash
- Clients reconnect; fetch latest doc state + recent ops from server
- Replay missed ops (or get snapshot if op log truncated)

### Network Partition
- CRDT: Continue editing; merge when partition heals
- OT: Queue ops; may need conflict resolution on reconnect

### Data Loss
- PostgreSQL: Replication, backups
- Operation log: Kafka with retention
- Version snapshots: Backup to cold storage

---

## 9. Monitoring & Observability

### Key Metrics
| Metric | Target |
|--------|--------|
| Op propagation latency p99 | < 100 ms |
| Document load time p99 | < 500 ms |
| WebSocket connection success | > 99.9% |
| Operation apply failure rate | < 0.01% |

### Logging
- Operation application errors
- Transformation failures (OT)
- Permission denials

### Tracing
- Trace op from client → server → other clients

---

## 10. Interview Tips

### Follow-up Questions
1. How would you implement copy-paste across documents?
2. How do you handle very long documents (10MB+)?
3. How would you add real-time collaboration to a spreadsheet?
4. How do you prevent abuse (e.g., one user spamming edits)?

### Common Mistakes
- Choosing OT vs CRDT without understanding trade-offs
- Ignoring presence and cursor sync
- Not considering offline
- Overlooking permission checks on every write

### What to Emphasize
- OT vs CRDT comparison
- Operation representation (insert, delete, retain)
- Transformation rules (OT)
- Presence and cursor tracking
- Version history and snapshots

---

## Appendix A: OT Transformation Rules (Detailed)

### Insert vs Insert
```
T(Ins(p1,c1), Ins(p2,c2)):
  if p1 < p2: return Ins(p2 + len(c1), c2)
  if p1 > p2: return Ins(p2, c2)
  if p1 == p2: return Ins(p2 + len(c1), c2)  // tie-break by client_id
```

### Insert vs Delete
```
T(Ins(p1,c1), Del(p2,len)):
  if p1 <= p2: return Del(p2 + len(c1), len)
  if p1 >= p2 + len: return Del(p2, len)
  else: // overlap - split delete or reject
```

### Delete vs Delete
```
T(Del(p1,len1), Del(p2,len2)):
  // Complex: handle overlapping ranges
  // Result: two deletes or one merged delete
```

### Implementation Notes
- Google's OT uses "operational transformation" with string model
- Etherpad uses OT with attributed string (formatting)
- Transformation must be correct for all operation pairs

---

## Appendix B: CRDT Types for Text

### RGA (Replicated Growable Array)
- Each character: (id, value, after_id)
- Id: (site_id, seq_num) — unique
- Insert: add after "after_id"
- Delete: tombstone
- Order: topological sort by (after_id, id)

### YATA (Yjs)
- Each character: (id, value, origin, left, right)
- Tree structure for efficient merge
- Used by Yjs library

### LSEQ
- Fractional indexing for order
- Avoids tombstones
- More complex merge

---

## Appendix C: Document Schema Example (ProseMirror)

```json
{
  "doc": {
    "content": "block+",
    "content": [
      {"type": "paragraph", "content": "inline*"},
      {"type": "heading", "attrs": {"level": 1}, "content": "inline*"},
      {"type": "bulletList", "content": "listItem+"},
      {"type": "orderedList", "content": "listItem+"}
    ]
  },
  "paragraph": {"content": "inline*", "group": "block"},
  "text": {"inline": true, "group": "inline"},
  "hardBreak": {"inline": true, "group": "inline"}
}
```

---

## Appendix D: WebSocket Scaling with Redis Pub/Sub

```
Client A (Server 1)          Redis Pub/Sub           Client B (Server 2)
     |                             |                        |
     |-- op ---------------------->|                        |
     |                             |-- PUBLISH doc:123 op ->|
     |                             |                        |-- op
     |                             |<-- SUBSCRIBE doc:123 ---|
     |                             |                        |
```

- Each WebSocket server subscribes to doc channels for connected clients
- On op: PUBLISH to doc channel
- All servers with subscribers receive; forward to their clients
- Horizontal scaling without sticky sessions per doc

---

## Appendix E: Conflict Resolution Examples

### OT: Concurrent Inserts
- Doc: "ab"
- A: Ins(1, "x") → "axb"
- B: Ins(2, "y") → "aby"
- Server receives A first: doc = "axb"
- B' = T(Ins(2,"y"), Ins(1,"x")) = Ins(3,"y") → "axby"
- Broadcast B' to A: A applies → "axby"
- Broadcast A to B: B transforms and applies → "axby"
- Converged

### CRDT: Same Scenario
- A: Ins after "a" with id (1,1)
- B: Ins after "b" with id (2,1)
- Merge: Union of both; order by (after_id, id)
- Result: "axby" or "abyx" depending on tie-break (deterministic)

---

## Appendix F: Permission Check Flow

```
Client                    API Gateway                 Document Service
   |                            |                            |
   |-- PATCH /doc/123 ---------->|                            |
   |                            |-- Auth token ---------------|
   |                            |-- Get user_id               |
   |                            |-- Check permission -------->|
   |                            |   (owner/editor/commenter)   |
   |                            |<-- Allow/Deny --------------|
   |<-- 200/403 ----------------|                            |
```

- Cache permission in session (doc_id, user_id) → role
- Invalidate on share/unshare

---

## Appendix G: Version Snapshot Strategy

### When to Snapshot
- Every 100 operations
- Every 5 minutes of inactivity
- On explicit "Save" (if supported)
- On document close (last editor leaves)

### Storage
- Full document content (JSON)
- Version number (monotonic)
- Timestamp
- Optional: Diff from previous (delta storage)

### Restore Flow
- User selects version from history UI
- Load snapshot for that version
- Create new version with snapshot content (branch) or replace current
- Notify other collaborators (if replace)

---

## Appendix H: Offline Sync (CRDT-Focused)

### When Offline
- Buffer operations in IndexedDB/localStorage
- Continue editing with local CRDT state
- Queue ops for sync

### When Back Online
- Send buffered ops to server
- Server merges (CRDT merge is commutative)
- Receive any ops from server (concurrent edits)
- Merge into local state
- Resolve any UI conflicts (rare with CRDT)

---

## Appendix I: Comment Anchoring

### Problem
- User adds comment to "paragraph 2"
- Another user edits paragraph 2 (insert/delete)
- Comment anchor must move with content

### Solution (OT)
- Anchor = (path, offset) in document model
- When ops apply: Transform anchor position
- T(anchor, op): update offset based on op

### Solution (CRDT)
- Anchor = (position_id, offset) in CRDT
- Position_id is stable; offset relative to that position
- On merge: Anchor stays with content

---

## Appendix J: Cursor Throttling

### Why
- Cursor moves on every keystroke
- 60+ updates/sec would overwhelm
- Throttle to 10/sec max

### Implementation
- Client: Debounce 100 ms before send
- Server: If same user, same doc: only forward latest
- Store: Redis with last cursor per user; overwrite
