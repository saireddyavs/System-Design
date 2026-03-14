# Design Slack

## 1. Problem Statement & Requirements

### Problem Statement
Design a team collaboration platform like Slack that enables real-time messaging in channels and DMs, threading, file sharing, search, workspace isolation, and integrations/bots.

### Functional Requirements
- **Real-time messaging**: Send/receive messages in channels and DMs; instant delivery
- **Channels**: Public/private channels; join/leave; @mentions
- **Threads**: Reply to messages in threads; thread view
- **Search**: Full-text search across all messages user has access to
- **Workspace isolation**: Each org has its own workspace; users belong to one or more
- **Permissions**: Channel-level (public/private); workspace admin
- **File sharing**: Upload, share files; preview images/docs
- **Integrations/bots**: Incoming webhooks; bot users; slash commands
- **Presence**: Online/away/offline; typing indicators

### Non-Functional Requirements
- **Scale**: 750K+ organizations, millions of concurrent users
- **Latency**: Message delivery < 100ms
- **Availability**: 99.99%
- **Consistency**: Eventually consistent for message order; strong for critical ops

### Out of Scope
- Video/voice calls (separate service)
- Enterprise grid (multi-workspace)
- Compliance hold/ediscovery (simplified)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Workspaces**: 750K
- **Users**: 50M total; 10M DAU
- **Concurrent connections**: 2M (WebSocket)
- **Messages/day**: 1B
- **Channels/workspace**: 50 avg
- **Messages/channel/day**: 100

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Message send | 1B | ~35,000 |
| Message read (fan-out) | 1B × 20 members ≈ 20B | ~700,000 (write ampl) |
| Search | 100M | ~3,500 |
| File upload | 50M | ~1,700 |
| Presence/typing | 2M × 1/min | ~35,000 |

### Storage (5 years)
- **Messages**: 1.8T × 500B ≈ 900 TB
- **Files**: 50M × 5MB avg × 5 years ≈ 1.25 PB
- **Metadata**: Channels, users, memberships ≈ 100 GB
- **Search index**: 1.8T × 1KB ≈ 1.8 PB (Elasticsearch)

### Bandwidth
- **Message fan-out**: 35K × 20 × 500B ≈ 350 MB/s
- **WebSocket**: 2M connections × 1 msg/10s × 500B ≈ 100 MB/s
- **File upload**: 1.7K × 5MB ≈ 8.5 GB/s

### Cache
- **Online users**: 2M × 100B ≈ 200 MB (Redis)
- **Channel members**: 750K × 50 × 20 × 8B ≈ 6 GB (Redis)
- **Recent messages**: 100K channels × 100 msgs × 500B ≈ 5 GB (Redis)

---

## 3. API Design

### REST Endpoints

```
# Workspaces & Auth
POST   /api/v1/auth/login
GET    /api/v1/workspaces/:workspace_id

# Channels
GET    /api/v1/channels
POST   /api/v1/channels
GET    /api/v1/channels/:channel_id
POST   /api/v1/channels/:channel_id/join
POST   /api/v1/channels/:channel_id/leave

# Messages
GET    /api/v1/channels/:channel_id/messages
Query: before, after, limit=50
POST   /api/v1/channels/:channel_id/messages
Body: { "text": "...", "thread_ts": "..." }

GET    /api/v1/channels/:channel_id/threads/:thread_ts/replies
POST   /api/v1/channels/:channel_id/threads/:thread_ts/replies
Body: { "text": "..." }

# DMs
GET    /api/v1/dm/conversations
POST   /api/v1/dm/conversations
Body: { "user_ids": ["u1", "u2"] }
GET    /api/v1/dm/conversations/:id/messages
POST   /api/v1/dm/conversations/:id/messages

# Search
GET    /api/v1/search
Query: q=query, channel_id, sort=relevance|timestamp

# Files
POST   /api/v1/files/upload
GET    /api/v1/files/:file_id
```

### WebSocket Endpoints

```
WS /ws/connect
Query: token, workspace_id

# Client subscribes to channels on connect
# Server pushes:
# - message.new
# - message.changed
# - message.deleted
# - typing.start, typing.stop
# - presence.change
# - channel.join, channel.leave
```

### WebSocket Message Format

```json
// Client → Server
{ "type": "message", "channel_id": "C123", "text": "Hello" }
{ "type": "typing", "channel_id": "C123" }

// Server → Client
{ "type": "message", "channel_id": "C123", "message": {...} }
{ "type": "typing", "user_id": "U456", "channel_id": "C123" }
{ "type": "presence", "user_id": "U456", "status": "active" }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Messages**: Cassandra (write-heavy, partition by channel)
- **Channels, users, memberships**: PostgreSQL
- **Search**: Elasticsearch
- **Presence, typing**: Redis (ephemeral)
- **Files**: S3 + metadata in PostgreSQL
- **WebSocket state**: In-memory + Redis (connection mapping)

### Schema

**Workspaces (PostgreSQL)**
```sql
workspaces (
  workspace_id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100),
  domain VARCHAR(100),
  created_at TIMESTAMP
)
```

**Channels (PostgreSQL)**
```sql
channels (
  channel_id VARCHAR(20) PRIMARY KEY,
  workspace_id BIGINT,
  name VARCHAR(100),
  is_private BOOLEAN,
  created_at TIMESTAMP
)

channel_members (
  channel_id VARCHAR(20),
  user_id VARCHAR(20),
  joined_at TIMESTAMP,
  PRIMARY KEY (channel_id, user_id)
)
```

**Messages (Cassandra)**
```sql
messages (
  channel_id VARCHAR(20),
  message_ts TIMESTAMP,      -- Unique per channel (snowflake-like)
  user_id VARCHAR(20),
  text TEXT,
  thread_ts TIMESTAMP,       -- NULL if top-level
  reply_count INT,
  created_at TIMESTAMP,
  PRIMARY KEY (channel_id, message_ts)
) WITH CLUSTERING ORDER BY (message_ts DESC);
```

**DMs (Cassandra)**
```sql
dm_conversations (
  dm_id VARCHAR(20) PRIMARY KEY,
  user_ids SET<VARCHAR>,     -- Sorted for consistency
  created_at TIMESTAMP
)

dm_messages (
  dm_id VARCHAR(20),
  message_ts TIMESTAMP,
  user_id VARCHAR(20),
  text TEXT,
  PRIMARY KEY (dm_id, message_ts)
) WITH CLUSTERING ORDER BY (message_ts DESC);
```

**Users (PostgreSQL)**
```sql
users (
  user_id VARCHAR(20) PRIMARY KEY,
  workspace_id BIGINT,
  email VARCHAR(255),
  name VARCHAR(100),
  avatar_url VARCHAR(500),
  created_at TIMESTAMP
)
```

**Files (PostgreSQL + S3)**
```sql
files (
  file_id VARCHAR(20) PRIMARY KEY,
  channel_id VARCHAR(20),
  user_id VARCHAR(20),
  filename VARCHAR(255),
  s3_key VARCHAR(500),
  size_bytes BIGINT,
  mime_type VARCHAR(100),
  created_at TIMESTAMP
)
```

**Presence (Redis)**
```
Key: presence:{user_id}
Value: { "status": "active", "last_seen": ts }
TTL: 5 minutes (refresh on activity)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENTS (Web, Mobile, Desktop)            │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                    ┌─────────────────────────────┴─────────────────────────────┐
                                    │                         HTTPS                              │
                                    └─────────────────────────────┬─────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              API GATEWAY / LOAD BALANCER                                       │
└───────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬───────────────────┐
        │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Message       │   │ Channel       │   │ Search        │   │ File          │
│ Service       │   │ Service       │   │ Service       │   │ Service       │
│               │   │               │   │               │   │               │
│ - Persist     │   │ - CRUD        │   │ - Elasticsearch│   │ - Upload      │
│ - Fan-out     │   │ - Members     │   │ - Index msgs  │   │ - S3          │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │                   │
        │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Cassandra     │   │ PostgreSQL    │   │ Elasticsearch │   │ S3            │
│ (Messages)    │   │ (Channels,    │   │ (Search)      │   │ (Files)       │
│               │   │  Users)       │   │               │   │               │
└───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘
        │                   │
        │                   │
        └───────────────────┼───────────────────────────────────────────────────┐
                            │                                                   │
                            ▼                                                   ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    WEBSOCKET LAYER                                                            │
│                                                                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                     │
│  │ WS Server 1    │    │ WS Server 2    │    │ WS Server N     │    │ Redis Pub/Sub    │                     │
│  │ (Connections)  │    │ (Connections)  │    │ (Connections)  │    │ (Fan-out)        │                     │
│  └────────┬───────┘    └────────┬───────┘    └────────┬───────┘    └────────┬───────┘                     │
│           │                    │                    │                    │                              │
│           └────────────────────┴────────────────────┴────────────────────┘                              │
│                                        │                                                                      │
│                                        │ Subscribe: channel:C123, user:U456                                 │
│                                        ▼                                                                      │
│                            Message Service publishes to channel:C123                                         │
│                            All WS servers subscribed push to connected clients                                │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Real-Time Messaging Flow

1. **User sends message** via WebSocket or REST
2. **Message Service**: Validate; persist to Cassandra
3. **Fan-out**: Publish to Redis channel `channel:{channel_id}`
4. **WebSocket servers**: Subscribed to Redis; receive message
5. **Push**: Each WS server pushes to connected clients in that channel
6. **Client**: Renders message; updates UI

### 6.2 WebSocket Connection Management

- **Sticky sessions**: Client must connect to same WS server (or use Redis Pub/Sub)
- **Connection mapping**: Redis `user:{id}:channels` → set of channel_ids; `channel:{id}:users` → set of user connections
- **Heartbeat**: Ping/pong every 30s; disconnect if no response
- **Reconnect**: Client reconnects; fetches missed messages via REST (before/after cursor)

### 6.3 Message Persistence

- **Write path**: Message Service writes to Cassandra (channel_id, message_ts, ...)
- **message_ts**: Snowflake or UUID; sortable
- **Threads**: thread_ts = parent message_ts; replies have same thread_ts
- **Read path**: Paginate by message_ts DESC

### 6.4 Search (Full-Text)

- **Index**: Elasticsearch; index each message with channel_id, user_id, workspace_id, text, timestamp
- **Access control**: Filter by workspace + channel membership at query time
- **Sync**: Kafka or CDC from Cassandra to Elasticsearch
- **Near real-time**: 1-5s delay acceptable

### 6.5 Presence & Typing Indicators

- **Presence**: Client sends heartbeat; Redis `presence:{user_id}` with TTL 5m
- **Broadcast**: On change, publish to user's channels/DMs
- **Typing**: Ephemeral; no persistence; Redis with 5s TTL; broadcast to channel
- **Away**: No heartbeat for 5m → status = away

### 6.6 File Sharing

- **Upload**: Multipart to API; store in S3; metadata in PostgreSQL
- **Share**: Post message with file_id; message contains file reference
- **Preview**: Generate thumbnails (images); store in S3
- **CDN**: Serve files via CDN

### 6.7 Integrations & Bots

- **Incoming webhook**: POST to URL; creates message as "bot"
- **Bot user**: OAuth; bot has user_id; can post, read
- **Slash commands**: `/remind me in 1h`; routed to bot service; bot responds
- **Event subscription**: Bot subscribes to channel events; receives via webhook

### 6.8 Workspace Isolation

- **Data**: All channels, users scoped by workspace_id
- **Auth**: JWT includes workspace_id; validate on every request
- **Search**: Filter by workspace + user's accessible channels

---

## 7. Scaling

### WebSocket Scaling
- **Horizontal**: Multiple WS servers; Redis Pub/Sub for fan-out
- **Connection limit**: ~10K connections per server (ephemeral)
- **2M connections**: 200 WS servers

### Message Fan-Out
- **Redis Pub/Sub**: O(1) per subscriber; scales to 100K subs per channel
- **Alternative**: Kafka per channel; consumers push to WS
- **Large channels**: 10K+ members; consider lazy load for read receipts

### Cassandra
- **Partition**: By channel_id; hot channels (large) may need partition splitting
- **Replication**: RF=3; multi-DC

### Elasticsearch
- **Sharding**: By workspace_id or channel_id
- **Index per workspace**: Isolate; easier deletion

---

## 8. Failure Handling

### WebSocket Disconnect
- Client reconnects; fetch messages since last seen (REST)
- Cursor: Last message_ts; `GET /messages?after=cursor`

### Message Service Down
- Queue messages in client; retry on reconnect
- Idempotency: Client sends idempotency key; server dedupes

### Cassandra Unavailable
- Circuit breaker; return cached recent messages
- Degrade: Read-only mode; queue writes

### Redis Pub/Sub Failure
- Fallback: Polling (high latency)
- Or: Direct fan-out from Message Service to WS servers (tight coupling)

---

## 9. Monitoring & Observability

### Key Metrics
- **Message latency**: Send to receive (p99)
- **WebSocket connections**: Count, reconnect rate
- **Search latency**: p50, p99
- **File upload**: Success rate, latency
- **API**: QPS, error rate per endpoint

### Alerts
- Message delivery latency > 500ms
- WebSocket disconnect rate > 5%
- Search p99 > 2s
- Cassandra write failure rate > 0.1%

### Tracing
- Trace: Message send → Persist → Pub/Sub → Push
- Correlation: message_id across services

---

## 10. Interview Tips

### Follow-up Questions
- "How do you handle a channel with 100K members?"
- "How would you implement message edit/delete with sync to all clients?"
- "How does search stay consistent with new messages?"
- "How would you add read receipts?"
- "How do you scale WebSocket connections to 10M?"

### Common Mistakes
- **No WebSocket**: REST polling doesn't scale for real-time
- **Ignoring fan-out**: 35K msg/s × 20 members = 700K deliveries
- **Single DB for messages**: Cassandra/Cassandra-like for write-heavy
- **No workspace isolation**: Critical for multi-tenant

### Key Points to Emphasize
- **Real-time**: WebSocket; Redis Pub/Sub for fan-out
- **Message persistence**: Cassandra; partition by channel
- **Search**: Elasticsearch; sync from Cassandra
- **Scale**: 750K workspaces, millions of concurrent users
- **Presence**: Redis; heartbeat; broadcast on change

---

## Appendix: Extended Design Details & Walkthrough Scenarios

### A. Message Fan-Out Walkthrough

1. User A sends "Hello" to channel C123
2. Message Service: Write to Cassandra; get message_ts
3. Publish to Redis: `PUBLISH channel:C123 '{"message_ts":..., "text":"Hello"}'`
4. WS Server 1, 2, 3 are subscribed to `channel:C123`
5. Each server has 100 clients in C123
6. Each server pushes to its 100 clients
7. Total: 300 clients receive message in < 100ms

### B. Thread Model

- **Top-level**: thread_ts = NULL
- **Reply**: thread_ts = parent message_ts
- **Query**: `WHERE channel_id = ? AND thread_ts = ?`
- **Reply count**: Denormalized on parent; increment on new reply

### C. Connection Mapping (Redis)

```
user:U123:connections → { ws_server_1:conn_1, ws_server_2:conn_2 }
channel:C123:members → { U1, U2, U3, ... }
ws_server_1:channel:C123 → set of connection IDs
```

On message: Get channel members → Get their connection locations → Publish to each WS server

### D. Search Index Schema (Elasticsearch)

```json
{
  "mappings": {
    "properties": {
      "message_id": { "type": "keyword" },
      "channel_id": { "type": "keyword" },
      "workspace_id": { "type": "keyword" },
      "user_id": { "type": "keyword" },
      "text": { "type": "text" },
      "message_ts": { "type": "date" },
      "thread_ts": { "type": "keyword" }
    }
  }
}
```

Query: Filter by workspace_id + channel_ids (user has access); full-text on text

### E. Slash Command Flow

1. User: `/remind me in 1h buy milk`
2. Message Service: Detects slash; routes to Bot Service
3. Bot Service: Parses; schedules reminder; responds "Reminder set for 3pm"
4. Bot Service posts response as bot user to same channel
5. Message flows through normal path

### F. File Upload Flow

1. Client: POST multipart to /files/upload
2. API: Validate (size, type); upload to S3; get key
3. Insert into files table; get file_id
4. Create message with file_id
5. Message contains: { "type": "file", "file_id": "F123" }
6. Clients render file preview/link

### G. Presence Broadcast

- User U1 goes online
- Client sends `presence.change` to WS
- WS server publishes to `user:U1:channels` (Redis)
- All channels U1 is in: C1, C2, C3
- Publish to `channel:C1`, `channel:C2`, `channel:C3`: "U1 is online"
- All members of C1, C2, C3 receive update

### H. Large Channel (100K members)

- **Fan-out**: 100K Redis subscribers; 100K pushes
- **Optimization**: Don't push to offline users; they fetch on reconnect
- **Read receipts**: Too expensive; skip or sample
- **Typing**: Broadcast to 100K; consider disabling or rate-limit
- **Alternative**: Shard channel; sub-channels for presence

### I. Message Ordering

- **Cassandra**: message_ts is sort key; chronological
- **Clock**: Use hybrid logical clock or snowflake for distributed ordering
- **Conflict**: Same millisecond; use client id as tiebreaker

### J. Workspace Migration

- Export: Dump channels, messages, users
- Import: Create new workspace; bulk insert
- Search: Reindex in new workspace
- Not in scope for MVP; mention for enterprise
