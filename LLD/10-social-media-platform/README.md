# Social Media Platform - Low Level Design

A production-quality Go implementation of a social media platform following clean architecture and SOLID principles. Interview-ready design (45-50 min discussion).

## Problem Description

Design and implement a social media platform with the following capabilities:

- **User Management**: Register, update profile, deactivate accounts
- **Social Graph**: Send/accept/reject friend requests, unfollow
- **Content**: Create, edit, delete posts (text + image URLs)
- **Engagement**: Comment on posts, like/unlike posts and comments
- **Feed**: News feed from friends/followed users with configurable sorting
- **Notifications**: Real-time notifications for likes, comments, friend requests
- **Discovery**: User search by username, email, or bio

## Requirements

| Requirement | Implementation |
|-------------|----------------|
| User profile management | `UserService` with Facade pattern |
| Friend/follow system | `FriendshipService` with request lifecycle |
| Create/edit/delete posts | `PostService` with authorization checks |
| Comments | `CommentRepository` + `PostService.AddComment` |
| Like/unlike | `LikeRepository` with post/comment support |
| News feed | `FeedService` with Strategy pattern |
| Notifications | Observer pattern via `NotificationPublisher` |
| User search | `UserRepository.Search` with fuzzy matching |

## Core Entities & Relationships

```
┌─────────┐     Friendship      ┌─────────┐
│  User   │◄──────────────────►│  User   │
└────┬────┘  (Pending/Accepted) └────┬────┘
     │                               │
     │ creates                       │ creates
     ▼                               ▼
┌─────────┐                    ┌──────────┐
│  Post   │◄─── Comment ────────│ Comment  │
└────┬────┘                    └──────────┘
     │
     ├── Like (many users)
     │
     └── triggers Notification
```

### Entity Details

| Entity | Key Fields | Relationships |
|--------|------------|---------------|
| **User** | ID, Username, Email, Bio, ProfilePicURL, IsActive | Friends (bidirectional), Followers, Following |
| **Post** | ID, AuthorID, Content, ImageURLs, CreatedAt | Author, Comments, Likes |
| **Comment** | ID, PostID, AuthorID, Content | Post, Author, Likes |
| **Friendship** | ID, RequesterID, ReceiverID, Status | Requester, Receiver |
| **Notification** | ID, UserID, Type, SourceUserID, TargetID, Read | User (recipient), Source (actor) |

## Feed Generation Algorithm

### Algorithm Steps

1. **Collect Author IDs**: Get all accepted friends for the user + include self
2. **Fetch Posts**: Query posts from `PostRepository.GetPostsByAuthors(authorIDs, limit, offset)`
3. **Enrich with Metrics**: For each post, get like count and comment count
4. **Apply Sort Strategy**: Use injected `FeedSortStrategy` (chronological or popularity)
5. **Paginate**: Return slice with offset and limit

### Sort Strategies

| Strategy | Implementation | Use Case |
|----------|----------------|----------|
| **Chronological** | Sort by `CreatedAt` descending | Default feed, "Latest" tab |
| **Popularity** | Sort by `(likes + comments)` descending, then timestamp | "Trending" or "Top" tab |

### Pseudocode

```go
func GetFeed(userID, limit, offset) {
    friends := GetAcceptedFriends(userID)
    authorIDs := append(friends, userID)
    posts := GetPostsByAuthors(authorIDs, limit+offset+100, 0)
    items := enrichWithEngagement(posts)
    sorted := strategy.Sort(items)
    return sorted[offset : offset+limit]
}
```

## Design Patterns

### 1. Observer Pattern (Notifications)

**Where**: `NotificationPublisher`, `NotificationService`, `PostService`, `FriendshipService`

**Why**: Decouples event producers (like, comment, friend request) from consumers (notification storage, push, email). When a like occurs, we publish a notification event—any subscriber can react (persist, send push, send email) without the producer knowing.

```go
// Publisher notifies all observers
publisher.Publish(notification)

// NotificationService persists; future: PushService could send push
```

### 2. Strategy Pattern (Feed Sorting)

**Where**: `FeedSortStrategy` interface, `ChronologicalFeedStrategy`, `PopularityFeedStrategy`, `FeedService`

**Why**: Feed sorting algorithms vary (chronological, popularity, ML-based). Strategy allows swapping algorithms at runtime without modifying `FeedService`. Open/Closed principle—add new strategies without changing existing code.

```go
feedService.SetStrategy(&PopularityFeedStrategy{})
feed, _ := feedService.GetFeed(userID, 10, 0)
```

### 3. Repository Pattern (Data Access)

**Where**: `UserRepository`, `PostRepository`, `FriendshipRepository`, `InMemory*Repository`

**Why**: Abstracts data storage. Services depend on interfaces, not concrete DB implementations. Easy to swap in-memory for PostgreSQL, Redis, etc. Enables unit testing with mocks.

### 4. Factory Pattern (Object Creation)

**Where**: `models.NewUser`, `models.NewPost`, `models.NewComment`, `models.NewFriendship`, `models.NewNotification`

**Why**: Centralizes object creation with consistent initialization (IDs, timestamps, default values). Encapsulates creation logic.

### 5. Facade Pattern (User Profile)

**Where**: `UserService` (Register, UpdateProfile, Deactivate)

**Why**: Profile operations involve validation, updates, and potentially multiple subsystems. Facade provides a simple interface: `UpdateProfile(userID, bio, pic)` hides internal complexity.

## SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **S - Single Responsibility** | Each service/repo has one job: `UserService` (users), `PostService` (posts/comments/likes), `FriendshipService` (friendships) |
| **O - Open/Closed** | `FeedSortStrategy`—open for new strategies, closed for modification of `FeedService` |
| **L - Liskov Substitution** | `InMemoryUserRepository` can replace any `UserRepository`; `ChronologicalFeedStrategy` can replace any `FeedSortStrategy` |
| **I - Interface Segregation** | Small, focused interfaces: `UserRepository`, `CommentRepository`, `LikeRepository` instead of one giant `DataStore` |
| **D - Dependency Inversion** | Services depend on `UserRepository` interface, not `InMemoryUserRepository`; high-level modules don't depend on low-level modules |

## Scalability Considerations

| Concern | Approach |
|---------|----------|
| **Read-heavy feed** | Cache feed per user (Redis), invalidate on new post from friends |
| **Write scaling** | Async notification publishing (message queue), eventual consistency for feed |
| **Friend list size** | Paginate `GetFriends`; for large graphs, use graph DB (Neo4j) or adjacency lists |
| **Feed generation** | Pre-compute feed in background job; use fan-out on write (store feed per user) or fan-out on read (current) |
| **Search** | Elasticsearch for user search; current in-memory search is for demo only |
| **Concurrency** | All repositories use `sync.RWMutex` for thread-safe access |

## Interview Explanations

### 3-Minute Summary

"We built a social media platform in Go with clean architecture. Core entities: User, Post, Comment, Friendship, Notification. We use the **Repository pattern** for data access—services depend on interfaces, so we can swap in-memory for DB. The **Observer pattern** handles notifications: when someone likes a post, we publish an event; the notification service persists it. The **Strategy pattern** powers feed sorting—chronological vs popularity—swappable at runtime. We follow SOLID: each service has one responsibility, we depend on abstractions, and new feed strategies extend without modifying existing code. The feed algorithm collects posts from friends, enriches with engagement metrics, applies the sort strategy, and paginates."

### 10-Minute Deep Dive

1. **Architecture**: Clean architecture—models, interfaces, services, repositories. No framework lock-in.

2. **Design Patterns**:
   - **Observer**: `NotificationPublisher` + subscribers. PostService/FriendshipService publish; NotificationService persists. Future: add PushService as subscriber.
   - **Strategy**: `FeedSortStrategy` interface. `ChronologicalFeedStrategy` and `PopularityFeedStrategy` implement it. `FeedService.SetStrategy()` allows runtime switch.
   - **Repository**: All data access behind interfaces. In-memory impl for demo; production would use PostgreSQL/Redis.
   - **Factory**: `models.NewPost`, etc. Centralize creation.
   - **Facade**: `UserService.UpdateProfile` simplifies profile updates.

3. **Feed Algorithm**: Get friends → fetch their posts → enrich with likes/comments → sort by strategy → paginate. Trade-off: fan-out on read (simple, good for small friend lists) vs fan-out on write (complex, better for large scale).

4. **SOLID**: SRP in services; OCP via Strategy; LSP with repository impls; ISP with small interfaces; DIP—services take interfaces in constructors.

5. **Thread Safety**: All in-memory repos use `sync.RWMutex`. Production would rely on DB transactions.

6. **Testing**: Unit tests for PostService (CRUD, likes, comments), FeedService (generation, strategies, pagination), FriendshipService (request lifecycle, notifications).

## Future Improvements

- [ ] **HTTP API**: Add REST/gRPC handlers with proper routing
- [ ] **Database**: PostgreSQL for persistence, Redis for caching
- [ ] **Authentication**: JWT, OAuth integration
- [ ] **Real-time**: WebSockets for live notifications
- [ ] **Media**: S3/Cloud Storage for image uploads (not just URLs)
- [ ] **ML Feed**: Personalized feed using user engagement history
- [ ] **Rate Limiting**: Protect APIs from abuse
- [ ] **Event Sourcing**: For audit trail and replay
- [ ] **GraphQL**: Flexible querying for mobile clients

## Running the Project

```bash
# Build
go build ./...

# Run demo
go run ./cmd/main.go

# Run tests
go test ./tests/... -v
```

## Directory Structure

```
10-social-media-platform/
├── cmd/main.go                 # Demo entry point
├── internal/
│   ├── models/                 # Domain entities
│   ├── interfaces/             # Repository & strategy contracts
│   ├── services/               # Business logic
│   └── repositories/            # In-memory implementations
├── tests/                      # Unit tests
├── go.mod
└── README.md
```
