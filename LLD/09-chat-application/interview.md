# Chat Application — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm 1:1 vs group, max members, real-time vs polling, offline handling |
| 2. Core Models | 7 min | User, Message, ChatRoom (with Members, Type), RoomMember |
| 3. Repository Interfaces | 5 min | UserRepository, MessageRepository, ChatRoomRepository |
| 4. Service Interfaces | 5 min | MessageBroker (Subscribe, Publish, Unsubscribe), MessageService, ChatService |
| 5. Core Service Implementation | 12 min | MessageService.SendMessage + Broker.Publish/deliverToUser (subscriber map, queue) |
| 6. Handler / main.go Wiring | 5 min | Wire repos, broker with DirectDelivery, MessageFactory |
| 7. Extend & Discuss | 8 min | Observer pattern, offline queue, DeliveryStrategy, concurrency |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Chat types: 1:1 and group? Max group size?
- Real-time: push to online users? How? (Channels, WebSocket)
- Offline: queue messages? Deliver when user comes online?
- Message history: paginated? Read receipts?
- Who can add/remove group members? (Creator, admin)

**Scope in:** User auth, 1:1 and group rooms, send message, real-time delivery via channels, offline queue, message history.

**Scope out:** Typing indicators, reactions, file attachments, E2E encryption.

## Phase 2: Core Models (7 min)

**Start with:** `Message` — ID, SenderID, RoomID, Content, Type, Status, Timestamp, ReadBy. Core unit of communication.

**Then:** `ChatRoom` — ID, Name, Type (OneOnOne/Group), Members (or RoomMember with UserID, Role), CreatedBy. `RoomMember` — UserID, Role (Creator/Admin/Member), JoinedAt.

**Then:** `User` — ID, Username, Email, PasswordHash, Status (Online/Offline), LastSeen.

**Skip for now:** MessageType enum details, ReadBy implementation.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `UserRepository`: Create, GetByID, GetByUsername, Update
- `MessageRepository`: Create, GetByID, GetByRoomID (with limit, offset for pagination)
- `ChatRoomRepository`: Create, GetByID, Update (for adding/removing members)

**Skip:** Separate RoomMemberRepository; can embed in ChatRoom or use ChatRoomRepo methods.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `MessageBroker`: Subscribe(userID) -> <-chan *Message, Unsubscribe(userID), Publish(message, recipientIDs), QueueForOffline, GetQueuedMessages
- `DeliveryStrategy`: Deliver(broker, message, recipientIDs) — DirectDelivery iterates recipients
- `MessageService`: SendMessage, GetMessageHistory, Subscribe, Unsubscribe
- `ChatService`: CreateOneOnOneRoom, CreateGroupRoom, AddMember, RemoveMember, LeaveRoom
- `MessageFactory`: Create(senderID, roomID, content, type) — ID, timestamp, defaults

**Key abstraction:** MessageBroker decouples send from delivery; Strategy for delivery algorithm.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `MessageService.SendMessage(senderID, roomID, content, replyToID)` — this is where the core logic lives.

**Algorithm:**
1. Validate sender is room member
2. Get room members (recipient IDs)
3. Create message via MessageFactory.Create(...)
4. Persist message
5. Call `broker.Publish(message, recipientIDs)`

**Broker.Publish** (with DirectDelivery):
- For each recipientID: `deliverToUser(recipientID, message)`

**deliverToUser:**
1. Lock; get ch := subscribers[userID]
2. Unlock (avoid holding during send)
3. If ch exists: `select { case ch <- message: return; default: }` — non-blocking; if full, fall through
4. If not subscribed or channel full: `QueueForOffline(userID, message)`

**Subscribe:** Create buffered channel (cap 100), put in subscribers[userID], return channel. **Unsubscribe:** Close channel, delete from subscribers.

**Offline:** QueueForOffline appends to queues[userID]. When user connects, call GetQueuedMessages, deliver, ClearQueue.

**Concurrency:** RWMutex on subscribers and queues; copy observer list before iterating to avoid deadlock.

## Phase 6: main.go Wiring (5 min)

```go
userRepo := NewInMemoryUserRepository()
msgRepo := NewInMemoryMessageRepository()
roomRepo := NewInMemoryChatRoomRepository()

msgBroker := NewInMemoryBroker(&DirectDelivery{})
msgFactory := NewMessageFactory()

authService := NewAuthService(userRepo)
chatService := NewChatService(roomRepo, userRepo)
msgService := NewMessageService(msgRepo, roomRepo, msgBroker, msgFactory)
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Observer:** MessageBroker — producers publish, subscribers receive via channels; decoupled
- **Strategy:** DeliveryStrategy — DirectDelivery; could add PriorityDelivery, RateLimitedDelivery
- **Factory:** MessageFactory — consistent message creation
- **Repository:** Swap for PostgreSQL; broker for Redis Pub/Sub in production

**Extensions:**
- Redis Pub/Sub for multi-instance deployment
- WebSocket layer: client connects → Subscribe(userID) → forward from channel to WebSocket
- Message queue (Kafka) for offline with persistence and replay

## Tips

- **Prioritize if low on time:** SendMessage + Publish + deliverToUser with subscriber map; skip offline queue details.
- **Common mistakes:** Blocking send (use select with default); deadlock when holding lock during channel send; not validating room membership.
- **What impresses:** Non-blocking delivery, offline queue, per-user channels, DeliveryStrategy for extensibility.
