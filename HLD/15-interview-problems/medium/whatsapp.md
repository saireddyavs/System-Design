# Design WhatsApp

## 1. Problem Statement & Requirements

### Problem Statement
Design a real-time messaging system like WhatsApp that supports 1-on-1 and group chats, message delivery guarantees, read receipts, online/offline status, media sharing, and end-to-end encryption.

### Functional Requirements
- **1-on-1 messaging**: Send/receive text messages between two users
- **Group messaging**: Groups up to 1000 members
- **Delivery guarantees**: Sent, delivered, read receipts
- **Presence**: Online/offline/typing status
- **Media sharing**: Images, videos, documents, voice notes
- **End-to-end encryption**: Messages encrypted client-side
- **Message history**: Sync across devices
- **Offline support**: Queue messages for offline users

### Non-Functional Requirements
- **Latency**: Message delivery < 100ms (online users)
- **Reliability**: No message loss, at-least-once delivery
- **Availability**: 99.99%
- **Scale**: 2B users, 100B messages/day

### Out of Scope
- Voice/video calls
- Status/stories (ephemeral)
- Business API
- Message editing

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Users**: 2 billion
- **DAU**: 1 billion
- **Messages/day**: 100 billion
- **Avg messages per chat**: 2 (1-on-1)
- **Group chats**: 20% of messages
- **Media**: 30% of messages

### QPS Calculation
| Operation | Daily Volume | QPS |
|-----------|--------------|-----|
| Message send | 100B | ~1.2M |
| Message receive | 100B | ~1.2M |
| Presence updates | 10B | ~120K |
| Media upload | 30B | ~350K |

### Storage (5 years)
- **Messages**: 100B × 365 × 5 × 500 bytes ≈ 91 PB
- **Media**: 30B × 365 × 5 × 2MB ≈ 110 EB (need dedup, compression)
- **Metadata**: Chat lists, etc. ≈ 10 TB

### Bandwidth
- **Text**: 100B × 500 bytes ≈ 50 TB/day
- **Media**: 30B × 2MB ≈ 60 EB/day (CDN cached)

### Connections
- **Concurrent WebSockets**: 500M (50% DAU online)
- **Connections per server**: 50K-100K (epoll)

---

## 3. API Design

### REST Endpoints (for non-realtime operations)

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
Body: { "phone": "+1234567890", "device_id": "..." }

GET    /api/v1/users/me
GET    /api/v1/users/:id/presence

POST   /api/v1/chats
Body: { "type": "direct|group", "participants": [...] }
Response: { "chat_id": "c123" }

GET    /api/v1/chats/:id/messages?before=&limit=50
Response: { "messages": [...], "has_more": true }

POST   /api/v1/media/upload
Body: multipart/form-data
Response: { "media_id": "m123", "url": "..." }
```

### WebSocket Protocol (Primary for messaging)

```
Client -> Server:
  SEND_MESSAGE: { chat_id, content, message_id, timestamp }
  TYPING: { chat_id }
  PRESENCE: { status: "online|away|offline" }

Server -> Client:
  MESSAGE: { message_id, chat_id, sender_id, content, timestamp }
  DELIVERED: { message_id, chat_id, delivered_to: [user_ids] }
  READ: { message_id, chat_id, read_by: user_id }
  PRESENCE_UPDATE: { user_id, status }
  TYPING_INDICATOR: { user_id, chat_id }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Messages**: Cassandra (write-heavy, time-series)
- **Chat metadata**: Cassandra
- **User sessions**: Redis (connection mapping)
- **Media metadata**: Cassandra
- **Media files**: S3 + CDN

### Schema

**Users (MySQL/Cassandra)**
```sql
users (
  user_id BIGINT PRIMARY KEY,
  phone VARCHAR(20) UNIQUE,
  name VARCHAR(100),
  profile_pic_url VARCHAR(500),
  created_at TIMESTAMP
)
```

**Chats (Cassandra)**
```sql
chats (
  chat_id UUID PRIMARY KEY,
  type TEXT,  -- direct, group
  name TEXT,  -- for groups
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

chat_participants (
  chat_id UUID,
  user_id BIGINT,
  role TEXT,  -- admin, member
  joined_at TIMESTAMP,
  last_read_message_id UUID,
  PRIMARY KEY (chat_id, user_id)
)

user_chats (
  user_id BIGINT,
  chat_id UUID,
  last_message_at TIMESTAMP,
  PRIMARY KEY (user_id, last_message_at)
) WITH CLUSTERING ORDER BY (last_message_at DESC);
```

**Messages (Cassandra)**
```sql
messages (
  chat_id UUID,
  message_id TIMEUUID,
  sender_id BIGINT,
  content TEXT,
  type TEXT,  -- text, image, video, document
  media_id UUID,
  encrypted_payload BLOB,  -- for E2E
  created_at TIMESTAMP,
  PRIMARY KEY (chat_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

**Message Status (Cassandra)**
```sql
message_delivery (
  message_id TIMEUUID,
  user_id BIGINT,
  status TEXT,  -- sent, delivered, read
  timestamp TIMESTAMP,
  PRIMARY KEY (message_id, user_id)
)
```

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │              CLIENT (Mobile/Web) - WebSocket                 │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    LOAD BALANCER (Sticky sessions for WebSocket)                               │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
                    ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
                    ▼                                         ▼                                         ▼
         ┌──────────────────┐                      ┌──────────────────┐                      ┌──────────────────┐
         │   Connection     │                      │    Message       │                      │   Presence       │
         │   Service        │                      │    Service       │                      │   Service        │
         │  ─────────────   │                      │  ─────────────   │                      │  ─────────────   │
         │  • WebSocket     │                      │  • Route msg     │                      │  • Heartbeat     │
         │  • Session mgmt  │                      │  • Store msg     │                      │  • Online status │
         │  • User->Conn    │                      │  • Fan-out       │                      │  • Typing        │
         └────────┬─────────┘                      └────────┬─────────┘                      └────────┬─────────┘
                  │                                         │                                         │
                  │                                         ▼                                         │
                  │                              ┌──────────────────┐                                 │
                  │                              │  Message Queue   │                                 │
                  │                              │  (Kafka)         │                                 │
                  │                              │  • message-sent  │                                 │
                  │                              │  • delivery-ack  │                                 │
                  │                              └────────┬────────┘                                 │
                  │                                         │                                         │
                  ▼                                         ▼                                         ▼
         ┌──────────────────┐                      ┌──────────────────┐                      ┌──────────────────┐
         │  Redis           │                      │  Cassandra       │                      │  Redis            │
         │  user:conn map   │                      │  Messages        │                      │  Presence state   │
         └──────────────────┘                      └──────────────────┘                      └──────────────────┘
                                                              │
                                                              ▼
                                                    ┌──────────────────┐
                                                    │  S3 + CDN        │
                                                    │  Media storage   │
                                                    └──────────────────┘
```

### Message Flow (ASCII Architecture)

```
    USER A (Online)              CONNECTION SVC           MESSAGE SVC              USER B (Online)
         │                            │                        │                         │
         │  Send message              │                        │                         │
         │──────────────────────────>│                        │                         │
         │                            │  Route to Message Svc  │                         │
         │                            │───────────────────────>│                         │
         │                            │                        │  Store in Cassandra     │
         │                            │                        │  Publish to Kafka        │
         │                            │                        │                         │
         │                            │                        │  Lookup B's connection  │
         │                            │                        │  (Redis: user->conn)     │
         │                            │                        │                         │
         │                            │                        │  Push via WebSocket      │
         │                            │                        │────────────────────────>│
         │                            │                        │                         │
         │                            │                        │  Delivered receipt      │
         │                            │                        │<────────────────────────│
         │                            │                        │  Store delivery status  │
         │  Delivered receipt        │                        │  Push to A               │
         │<──────────────────────────│<───────────────────────│                         │
         │                            │                        │                         │


    USER B (Offline)               MESSAGE SVC              OFFLINE QUEUE            WHEN B COMES ONLINE
         │                            │                        │                         │
         │  B is offline               │                        │                         │
         │                            │  Store in Cassandra     │                         │
         │                            │  Push to offline queue  │                         │
         │                            │───────────────────────>│                         │
         │                            │                        │  Queue per user          │                         │
         │                            │                        │  (Kafka/RabbitMQ)       │                         │
         │                            │                        │                         │
         │                            │                        │  B connects             │
         │                            │                        │<────────────────────────│
         │                            │                        │  Consume queued msgs    │                         │
         │                            │                        │  Push to B's connection │                         │
         │                            │                        │────────────────────────>│
         │                            │                        │  Send delivered receipt│                         │
```

---

## 6. Detailed Component Design

### 6.1 WebSocket Connection Management
**Connection Service**:
- **One WebSocket per device**: user_id + device_id
- **Sticky sessions**: Load balancer routes by user_id to same server
- **Connection mapping**: Redis `user:{user_id}:conn` → server_id
- **Scale**: 50K connections/server (epoll, async I/O)
- **Reconnection**: Client reconnects, fetches missed messages via sync API

**Connection lifecycle**:
1. Client connects → authenticate (JWT)
2. Register in Redis: user_id → (server_id, connection_id)
3. On disconnect: remove from Redis, queue undelivered messages

### 6.2 Message Storage (Write-Ahead)
**Flow**:
1. **Persist first**: Write to Cassandra before acknowledging
2. **Then deliver**: Push to online recipients
3. **Offline**: Add to per-user queue (Kafka partition by user_id)
4. **Store until delivered**: Don't delete from queue until ack

**Ordering**: message_id (TIMEUUID) provides total order per chat

### 6.3 Message Queue for Offline Users
- **Kafka**: Topic `offline-messages`, partition by recipient user_id
- **Consumer**: Connection service consumes when user connects
- **Retention**: 30 days for offline queue
- **Deduplication**: message_id (user may connect from multiple devices)

### 6.4 Group Message Fan-out
**Option A: Fan-out on write**
- On send: push to all 1000 members' connections/queues
- Pros: Fast for online users
- Cons: 1000 writes per message

**Option B: Fan-out on read**
- Store once, each user fetches
- Pros: 1 write
- Cons: 1000 reads when loading chat

**Hybrid**: 
- Store in Cassandra (1 write)
- Push to online members only (real-time)
- Offline: each has partition in Kafka, consumed when online
- On chat open: fetch from Cassandra (paginated)

### 6.5 Presence Service (Heartbeat-based)
- **Heartbeat**: Client sends every 30s
- **Redis**: `presence:{user_id}` → {status, last_seen} (TTL 90s)
- **Status**: online (heartbeat < 90s), away (90-300s), offline (>300s)
- **Broadcast**: On status change, notify relevant contacts (people in chat)
- **Typing**: Ephemeral, 5s TTL, don't persist

### 6.6 Media Upload/Download
- **Upload**: Client → API → S3 (pre-signed URL for direct upload)
- **Metadata**: Store in Cassandra (media_id, url, size, type)
- **Download**: S3 + CDN (CloudFront)
- **Encryption**: E2E encrypted, store ciphertext

### 6.7 End-to-End Encryption (Signal Protocol)
- **Key exchange**: X3DH (Extended Triple Diffie-Hellman)
- **Ratchet**: Double Ratchet for forward secrecy
- **Server**: Never sees plaintext, stores encrypted payload
- **Key storage**: Client holds keys, server stores identity keys for discovery
- **Group**: Sender keys (each member has key, sender encrypts for each)

**Flow**:
1. A and B register identity keys with server
2. A fetches B's prekeys, performs X3DH, gets shared secret
3. A encrypts message with derived key, sends ciphertext
4. Server stores and forwards ciphertext
5. B decrypts with same derived key

---

## 7. Scaling

### Sharding
- **Messages**: Shard by chat_id
- **User chats**: Shard by user_id
- **Connections**: Shard by user_id (consistent hashing to server)

### Caching
- **Chat list**: Redis, invalidate on new message
- **Presence**: Redis (already)
- **Media**: CDN

### Connection Scaling
- **Horizontal**: Add more connection servers
- **Redis cluster**: For connection mapping
- **Kafka partitions**: Match connection server count

---

## 8. Failure Handling

### Message Delivery
- **At-least-once**: Retry until ack
- **Idempotency**: message_id dedup at receiver
- **Ordering**: Kafka partition per user preserves order

### Connection Failures
- **Server crash**: Clients reconnect to new server, sync missed messages
- **Redis failover**: Rebuild connection map from active connections

### Component Redundancy
- **Cassandra**: RF=3, quorum writes
- **Kafka**: Replication factor 3
- **Redis**: Cluster mode

---

## 9. Monitoring & Observability

### Key Metrics
- **Message latency**: Send to receive (p99)
- **Connection count**: Per server, total
- **Offline queue depth**: Per user, total
- **Delivery rate**: % messages delivered
- **WebSocket errors**: Disconnects, timeouts

### Alerts
- Message latency p99 > 500ms
- Offline queue depth > 1M
- Connection server CPU > 80%

---

## 10. Interview Tips

### Follow-up Questions
1. **How to handle 1000-member group?** Fan-out to online only, queue for offline
2. **Multi-device sync?** Each device has connection, all receive, last_read per device
3. **Message ordering across devices?** Use server timestamp (TIMEUUID)
4. **How to prevent replay attacks in E2E?** Include message_id in encryption, reject duplicates

### Common Mistakes
- **Synchronous delivery**: Must use queue for offline
- **No idempotency**: Duplicate messages on retry
- **Single connection server**: Can't scale
- **Storing plaintext**: E2E means server never decrypts

### Key Points
1. **WebSocket** for real-time, **store-and-forward** for reliability
2. **Redis** for connection mapping (user → server)
3. **Kafka** for offline queue
4. **Write-ahead** persist before deliver
5. **E2E** = server is dumb pipe, client does crypto

---

## Appendix A: Extended Design Details

### A.1 WebSocket Frame Format (Custom Protocol)
```
| type (1B) | length (4B) | payload |
type: 0=message, 1=delivered, 2=read, 3=typing, 4=presence
```

### A.2 Connection Server Capacity
- **epoll**: 50K connections per process (Linux)
- **Memory**: ~10KB per connection = 500MB per 50K
- **CPU**: Idle connections minimal, burst on message flood
- **Scale**: 500M connections / 50K = 10,000 connection servers

### A.3 Kafka Partition Strategy for Offline
- **Partition key**: user_id (recipient)
- **Partitions**: 1000 (match consumer count)
- **Ordering**: Per-user ordering guaranteed
- **Consumer**: Connection service subscribes, on user connect fetches partition

### A.4 E2E Encryption - Sender Keys (Groups)
- Each member has ratchet with sender
- Sender encrypts once per recipient (N encryptions for N members)
- Alternative: Sender keys - 1 encryption, each recipient has sender's key
- WhatsApp uses sender keys for efficiency

### A.5 Message Delivery State Machine
```
SENT (by sender) -> DELIVERED (received by recipient device) -> READ (recipient opened chat)
Store each transition with timestamp for receipts
```

### A.6 Media Message Flow
```
1. Client encrypts media, uploads to S3 (pre-signed URL)
2. Message contains media_id, not content
3. Recipient fetches media by media_id
4. Decrypt client-side
5. Cache decrypted locally (optional)
```

### A.7 Load Balancer Sticky Session Config
```
Hash: user_id (consistent hashing)
Timeout: 24h (WebSocket long-lived)
Fallback: If server down, rehash to new server
```

### A.8 Sample Message Payload (E2E)
```json
{
  "message_id": "uuid",
  "chat_id": "uuid",
  "sender_id": 123,
  "encrypted_content": "base64_ciphertext",
  "type": "text",
  "created_at": 1734567890
}
```

---

## Appendix B: Multi-Device Sync Deep Dive

### B.1 Device Registration
- Each device: (user_id, device_id) unique
- Server stores: device_id -> fcm_token (for push), capabilities
- Message delivery: Fan-out to all user's devices

### B.2 Last Read Per Device
```
chat_participants: (chat_id, user_id, device_id, last_read_message_id)
On "read" receipt: Update for specific device
"Read by" display: Show when all devices have read (or primary device)
```

### B.3 Message Sync on New Device
- Client connects, sends last_message_id from local storage
- Server: SELECT * FROM messages WHERE chat_id=? AND message_id > ?
- Paginated, return in batches
- Client merges with local, resolves conflicts by message_id

### B.4 Offline Queue Consumption
- User connects -> Connection service gets user_id
- Query: Which Kafka partition has this user_id? (hash(user_id) % partitions)
- Consumer: Fetch from partition, push to WebSocket
- Mark delivered: After successful push, commit offset
- Dedupe: Client sends ack with message_ids, server tracks

---

## Appendix C: Walkthrough Scenarios

### C.1 Scenario: A Sends Message to B (Both Online)
1. A's client sends SEND_MESSAGE via WebSocket
2. Connection server receives, routes to Message Service
3. Message Service: Validate chat, persist to Cassandra
4. Lookup B's connection: Redis user:456 -> server_2, conn_xyz
5. Message Service pushes to server_2's connection for user 456
6. B receives MESSAGE frame, displays in chat
7. B's client sends DELIVERED receipt
8. Message Service updates message_delivery table
9. Pushes DELIVERED to A's connection
10. A sees "Delivered" checkmark
11. End-to-end: < 100ms

### C.2 Scenario: A Sends Message to B (B Offline)
1. Steps 1-3 same (persist to Cassandra)
2. Lookup B's connection: Not found in Redis
3. Message Service publishes to Kafka offline-messages, partition=hash(456)
4. Returns to A: "Sent" (single checkmark)
5. 2 hours later: B opens app, establishes WebSocket
6. Connection service registers B in Redis
7. Offline consumer (or connection service) checks partition for user 456
8. Finds queued message, pushes to B's WebSocket
9. B receives, sends DELIVERED
10. Message Service pushes DELIVERED to A (if A online)
11. A sees "Delivered" (double checkmark)

### C.3 Scenario: Group Message (500 members, 200 online)
1. A sends to group chat
2. Message Service persists once to Cassandra
3. Fan-out: Get 500 member IDs
4. For each: Check Redis for connection
5. 200 online: Push directly to their WebSockets (parallel)
6. 300 offline: Publish 300 messages to Kafka (partition by recipient)
7. Each offline user gets message when they connect
8. Delivery receipts: Store per-recipient, send to A as they arrive

### C.4 Scenario: E2E Encrypted Message
1. A has shared secret with B (from X3DH key exchange)
2. A encrypts "Hello" with derived key, gets ciphertext
3. Sends {message_id, encrypted_content} to server
4. Server stores ciphertext, never decrypts
5. Forwards to B's device(s)
6. B decrypts with same key, displays "Hello"
7. Server is "dumb pipe" - cannot read message content

---

## Appendix D: Technology Choices Rationale

### D.1 Why WebSocket over HTTP Polling?
- **Latency**: Push vs poll (100ms vs 1-5s)
- **Efficiency**: One connection vs repeated requests
- **Battery**: Less radio usage on mobile
- **Real-time**: Typing indicators, presence
- **Trade-off**: Connection state, harder to scale

### D.2 Why Cassandra for Messages?
- **Write path**: Append-only, partition by chat_id
- **Ordering**: TIMEUUID gives total order per chat
- **Scale**: 1.2M writes/sec, Cassandra handles
- **No joins**: Denormalized, fetch by chat_id
- **TTL**: Optional for ephemeral messages

### D.3 Why Kafka for Offline Queue?
- **Partitioning**: By user_id for per-user ordering
- **Retention**: 30 days, replay if needed
- **Consumer groups**: Scale consumers
- **Durability**: Replicated, no loss
- **Decoupling**: Message service doesn't track connections

### D.4 Why Redis for Connection Map?
- **Speed**: Sub-ms lookup for routing
- **Ephemeral**: Connections come and go
- **Atomic**: CAS for connection registration
- **TTL**: Auto-cleanup on server crash

### D.5 Interview Discussion Points
- **E2E is non-negotiable**: Privacy requirement, server can't decrypt
- **Store-and-forward**: Reliability over real-time for offline users
- **WebSocket scale**: 500M connections = 10K servers at 50K each

---

## Quick Reference Summary

| Component | Technology | Key Reason |
|-----------|------------|------------|
| Connections | WebSocket + Redis | Real-time, routing |
| Messages | Cassandra | Write-heavy, ordered |
| Offline queue | Kafka | Per-user partition |
| Presence | Redis | Heartbeat, TTL |
| Media | S3 + CDN | Scale |
| E2E | Signal Protocol | Client-side crypto |
