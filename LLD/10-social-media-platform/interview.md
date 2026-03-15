# Social Media Platform — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm friends vs followers, feed source, sorting, notifications |
| 2. Core Models | 7 min | User, Post, Comment, Friendship, Notification, Like |
| 3. Repository Interfaces | 5 min | UserRepository, PostRepository, FriendshipRepository (GetAcceptedFriends), CommentRepository, LikeRepository |
| 4. Service Interfaces | 5 min | FeedService with FeedSortStrategy, PostService, FriendshipService, NotificationPublisher |
| 5. Core Service Implementation | 12 min | FeedService.GetFeed (friends → posts → enrich → sort → paginate) |
| 6. Handler / main.go Wiring | 5 min | Wire repos, FeedSortStrategy, NotificationPublisher with observers |
| 7. Extend & Discuss | 8 min | Adjacency list for friends, fan-out on read, Observer for notifications |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Social graph: friends (bidirectional) or followers (unidirectional)?
- Feed: posts from friends? Sorted how? (Chronological, popularity)
- Notifications: for likes, comments, friend requests?
- Posts: text + images (URLs)? Comments? Likes on posts and comments?
- User search: by username, email?

**Scope in:** User registration, friend requests (send/accept/reject), posts with comments and likes, feed from friends with sort strategies, notifications.

**Scope out:** Stories, DMs, hashtags, trending.

## Phase 2: Core Models (7 min)

**Start with:** `Post` — ID, AuthorID, Content, ImageURLs, CreatedAt. Core content unit.

**Then:** `Friendship` — ID, RequesterID, ReceiverID, Status (Pending/Accepted/Rejected). Bi-directional: when Accepted, both are friends. `User` — ID, Username, Email, Bio, IsActive.

**Then:** `Comment` — ID, PostID, AuthorID, Content. `Like` — PostID or CommentID, UserID (or separate LikeRepository with GetPostLikeCount). `Notification` — ID, UserID, Type, SourceUserID, TargetID, Read.

**Skip for now:** Profile pic URL, CreatedAt/UpdatedAt on Friendship.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `UserRepository`: Create, GetByID, Update, Search
- `PostRepository`: Create, GetByID, Update, Delete, **GetPostsByAuthors(authorIDs, limit, offset)**
- `FriendshipRepository`: Create, GetByID, Update, **GetAcceptedFriends(userID)** -> []string (friend IDs)
- `CommentRepository`: Create, GetByPostID
- `LikeRepository`: Create, Delete, GetPostLikeCount
- `NotificationRepository`: Create, GetByUserID (paginated)

**Skip:** GetFollowers/GetFollowing initially; GetAcceptedFriends is the key for feed.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `FeedSortStrategy`: Sort(items []FeedItem) []FeedItem — ChronologicalFeedStrategy, PopularityFeedStrategy
- `FeedService`: GetFeed(userID, limit, offset), SetStrategy(strategy)
- `PostService`: CreatePost, AddComment, LikePost, UnlikePost
- `FriendshipService`: SendFriendRequest, AcceptFriendRequest, RejectFriendRequest
- `NotificationPublisher`: Publish(notification), Subscribe(observer)
- `NotificationService`: GetUserNotifications (implements observer, persists)

**Key abstraction:** FeedSortStrategy is swappable; Observer for notifications.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `FeedService.GetFeed(userID, limit, offset)` — this is where the core logic lives.

**Algorithm:**
1. `friends := friendshipRepo.GetAcceptedFriends(userID)` — returns []string of friend IDs
2. Build authorIDs = {userID} ∪ friends (include self)
3. `posts := postRepo.GetPostsByAuthors(authorIDs, limit+offset+100, 0)` — fetch extra for sorting
4. For each post: get likeCount, commentCount, authorName; build FeedItem
5. `sorted := strategy.Sort(items)` — Chronological (by CreatedAt desc) or Popularity (likes+comments desc, then CreatedAt)
6. Paginate: return sorted[offset : offset+limit]

**GetAcceptedFriends (adjacency list):** Iterate friendships where Status==Accepted; if RequesterID==userID add ReceiverID, if ReceiverID==userID add RequesterID; return unique set. Bi-directional.

**Notification fan-out:** When PostService.LikePost or AddComment, create Notification, call publisher.Publish(notification). NotificationService (observer) persists. Future: PushService as another observer.

**Concurrency:** RWMutex on all repos; copy observers under lock before notifying.

## Phase 6: main.go Wiring (5 min)

```go
userRepo := NewInMemoryUserRepository()
friendshipRepo := NewInMemoryFriendshipRepository()
postRepo := NewInMemoryPostRepository()
commentRepo := NewInMemoryCommentRepository()
likeRepo := NewInMemoryLikeRepository()
notifRepo := NewInMemoryNotificationRepository()

publisher := NewNotificationPublisher()
notifService := NewNotificationService(notifRepo, publisher)
publisher.Subscribe(notifService)  // or notifService registers as observer

postService := NewPostService(postRepo, commentRepo, likeRepo, userRepo, notifService)
friendshipService := NewFriendshipService(userRepo, friendshipRepo, notifService, publisher)

chronologicalStrategy := &ChronologicalFeedStrategy{}
feedService := NewFeedService(postRepo, commentRepo, likeRepo, userRepo, friendshipRepo, chronologicalStrategy)
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Strategy:** FeedSortStrategy — chronological vs popularity; SetStrategy at runtime
- **Observer:** NotificationPublisher — like/comment/friend request publish; NotificationService persists
- **Repository:** Swap for PostgreSQL; cache feed in Redis
- **Facade:** UserService.UpdateProfile

**Extensions:**
- Fan-out on write: pre-compute feed per user when friend posts; faster read, more write load
- Graph DB (Neo4j) for large friend graphs
- ML-based personalized feed

## Tips

- **Prioritize if low on time:** GetFeed algorithm (friends → posts → sort → paginate); GetAcceptedFriends bi-directional logic.
- **Common mistakes:** Forgetting to include self in feed authors; wrong sort (asc vs desc); not copying observers before notify (concurrency).
- **What impresses:** Clean GetAcceptedFriends for bi-directional graph, Strategy for feed sort, Observer for notifications, fan-out on read vs write discussion.
