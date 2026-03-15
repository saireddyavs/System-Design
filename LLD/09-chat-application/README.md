# Chat Application - Low Level Design (LLD)

A production-quality, interview-ready Go implementation of a chat application following clean architecture and SOLID principles.

## 1. Problem Description & Requirements

### Problem
Design and implement a real-time chat application supporting user management, one-on-one messaging, group chats, message persistence, and online status tracking.

### Functional Requirements
- **User registration and authentication** - Secure signup/login with password hashing
- **One-on-one messaging** - Direct chat between two users
- **Group chat rooms** - Create, join, leave group chats (max 100 members)
- **Real-time message delivery** - Online users receive messages immediately via channels
- **Message persistence and history** - Paginated message retrieval
- **Online/offline status** - Track user presence and last seen
- **Message read receipts** - Track which users have read messages
- **User search** - Search by username or email

### Business Rules
- Group chat max 100 members
- Only creator or admin can add/remove members
- Messages delivered to online users immediately; queued for offline users
- Message history retrievable with pagination

---

## 2. Core Entities & Relationships

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    User     │       │   Message   │       │  ChatRoom   │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ ID          │       │ ID          │       │ ID          │
│ Username    │       │ SenderID ───┼───────│ Members[]   │
│ Email       │       │ RoomID ─────┼───────│ Type        │
│ PasswordHash│       │ Content     │       │ Name        │
│ Status      │       │ Type        │       │ CreatedBy   │
│ LastSeen    │       │ Status      │       └─────────────┘
│ CreatedAt   │       │ ReadBy[]    │
└─────────────┘       │ Timestamp   │
                      └─────────────┘

User 1───* RoomMember *───1 ChatRoom
Message *───1 ChatRoom
Message 1───1 User (Sender)
```

### Entity Details

| Entity | Key Fields | Relationships |
|--------|------------|---------------|
| **User** | ID, Username, Email, PasswordHash, Status, LastSeen | Sends messages, member of rooms |
| **Message** | ID, SenderID, RoomID, Content, Type, Status, ReadBy | Belongs to room, sent by user |
| **ChatRoom** | ID, Name, Type (OneOnOne/Group), Members, CreatedBy | Contains messages, has members |
| **RoomMember** | UserID, Role (Creator/Admin/Member), JoinedAt | Links User to ChatRoom |

---

## 3. Real-Time Delivery Mechanism

### Architecture
```
Sender → MessageService → Persist → MessageBroker.Publish()
                                    │
                                    ├─→ Online User A: channel ← goroutine receives
                                    ├─→ Online User B: channel ← goroutine receives
                                    └─→ Offline User C: queue (delivered when online)
```

### How It Works
1. **Subscribe**: When a user connects, they call `Subscribe(userID)` which creates a buffered channel (capacity 100) and registers it in the broker's subscriber map.
2. **Publish**: When a message is sent, the broker iterates recipient IDs. For each:
   - If **subscribed**: Non-blocking send to their channel (`select` with `default` to avoid blocking)
   - If **channel full** or **not subscribed**: Message is queued for offline delivery
3. **Receive**: Client goroutine reads from `<-chan *Message` until channel is closed.
4. **Unsubscribe**: On disconnect, `Unsubscribe(userID)` closes the channel and removes from map.
5. **Offline Queue**: `GetQueuedMessages(userID)` returns queued messages when user comes online; `ClearQueue(userID)` after delivery.

### Concurrency Model
- **Per-user channels**: Each subscriber has a dedicated channel; no shared channel contention
- **Mutex for shared state**: `subscribers` map and `queues` map protected by `sync.RWMutex`
- **Non-blocking send**: Prevents slow consumers from blocking the broker
- **Goroutine per consumer**: Each client runs a goroutine reading from its channel

---

## 4. Design Patterns

### Observer Pattern
**Where**: `MessageBroker` (subscribers receive messages)
**Why**: Decouples message producers from consumers. Multiple subscribers can receive the same message without the sender knowing who's listening. Enables real-time push without polling.

### Mediator Pattern
**Where**: `ChatService` (mediates room operations)
**Why**: ChatRoom acts as mediator between members—adding/removing members, routing messages. No direct member-to-member coupling for room management.

### Strategy Pattern
**Where**: `DeliveryStrategy` (DirectDelivery)
**Why**: Swappable delivery algorithms. Can add new strategies (e.g., priority delivery, rate-limited) without changing broker code. Open/Closed Principle.

### Repository Pattern
**Where**: `UserRepository`, `MessageRepository`, `ChatRoomRepository`
**Why**: Abstracts data access. Business logic doesn't know about storage (in-memory vs DB). Easy to swap implementations for testing or production.

### Factory Pattern
**Where**: `MessageFactory.Create()`
**Why**: Centralizes message creation logic (ID generation, default values, timestamps). Ensures consistent message structure across the codebase.

---

## 5. SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | `AuthService` (auth only), `UserService` (user ops), `MessageService` (messaging), `ChatService` (rooms). Each repository handles one entity type. |
| **O - Open/Closed** | `DeliveryStrategy` interface (DirectDelivery)—add new strategies without modifying broker. Repository interfaces allow new implementations. |
| **L - Liskov Substitution** | `InMemoryBroker` substitutes for `MessageBroker`; any `Repository` impl can substitute its interface. |
| **I - Interface Segregation** | Small, focused interfaces: `UserRepository`, `MessageRepository`, `ChatRoomRepository`, `MessageBroker`. No fat interfaces. |
| **D - Dependency Inversion** | Services depend on abstractions (interfaces), not concrete implementations. `main.go` wires concrete repos. |

---

## 6. Concurrency Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Shared State (Mutex)                      │
├─────────────────────────────────────────────────────────────┤
│ Repositories: map[string]*Entity  → RWMutex                  │
│ Broker: subscribers map, queues map → RWMutex                 │
└─────────────────────────────────────────────────────────────┘
         │
         │  Goroutines
         ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ User A channel  │  │ User B channel  │  │ User C channel  │
│ (goroutine)     │  │ (goroutine)     │  │ (goroutine)     │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

- **Mutex for repositories**: All map access protected; read-heavy use RLock for reads
- **Broker**: Lock for subscribe/unsubscribe/queue; no lock during Publish (deliverToUser acquires its own lock to avoid deadlock with QueueForOffline)
- **Channels**: Buffered (100) to absorb bursts; non-blocking send when full
- **Concurrent sends**: Tested with 10 goroutines sending 20 messages; thread-safe

---

## 7. Interview Explanations

### 3-Minute Pitch
"This is a chat application with clean architecture. Users register and authenticate. We support 1:1 and group chats with a 100-member limit. Messages are persisted and delivered in real-time via an Observer pattern—subscribers get a channel when they connect, and the broker sends messages to those channels. Offline users get messages queued. We use Repository pattern for data access, Strategy for delivery (DirectDelivery), and Factory for message creation. All shared state is protected by mutexes; each user has their own channel for lock-free delivery. SOLID principles are applied throughout—services depend on interfaces, not implementations."

### 10-Minute Deep Dive
1. **Architecture**: Clean architecture with models, interfaces, services, repositories. Dependency injection in main.
2. **Real-time**: MessageBroker implements Observer. Subscribe returns `<-chan *Message`. Publish uses DirectDelivery strategy. Online users get immediate delivery; offline queue.
3. **Concurrency**: RWMutex for repositories; broker uses per-user channels; non-blocking send in broker to avoid blocking on slow consumers.
4. **Business rules**: ChatService enforces add/remove (creator/admin only), max 100 members, leave room. MessageService validates room membership before send.
5. **Testing**: Unit tests for auth (register, login, duplicates), chat (create room, add member, leave, permissions), messaging (send, history, read receipts, real-time, concurrent sends).
6. **Extensibility**: Swap MessageBroker for Redis pub/sub; swap repositories for PostgreSQL; add WebSocket layer that subscribes and forwards to clients.

---

## 8. Future Improvements

1. **Persistence**: Replace in-memory repos with PostgreSQL/MongoDB
2. **Real-time transport**: WebSocket/SSE layer that subscribes to broker and pushes to clients
3. **Distributed broker**: Redis Pub/Sub or RabbitMQ for multi-instance deployment
4. **Rate limiting**: Per-user message rate limits
5. **Typing indicators**: Extend broker with typing events
6. **Message reactions**: Extend Message model with reactions map
7. **File attachments**: S3 for file/image message types (extend MessageType)
8. **API layer**: HTTP/gRPC handlers with middleware (auth, logging)
9. **Caching**: Redis for online status, recent messages
10. **Metrics**: Prometheus for message throughput, latency

---

## 9. Running the Application

```bash
# Install dependencies
go mod tidy

# Run the demo
go run ./cmd/main.go

# Run tests
go test ./tests/... -v
```

## 10. Directory Structure

```
09-chat-application/
├── cmd/main.go                 # Entry point, demo
├── internal/
│   ├── models/                 # User, Message, ChatRoom, Enums
│   ├── interfaces/             # Repository & Broker interfaces
│   ├── services/               # Auth, User, Chat, Message, MessageFactory
│   ├── repositories/           # In-memory implementations
│   ├── broker/                 # In-memory broker (Observer + Strategy)
│   └── apperrors/              # Shared errors
├── tests/                      # Unit tests
├── go.mod
└── README.md
```
