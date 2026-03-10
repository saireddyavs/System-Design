package services

import (
	"social-media-platform/internal/interfaces"
)

// FeedService generates the news feed for users.
// Strategy: Uses feed sort strategy for ordering (chronological vs popularity)
type FeedService struct {
	postRepo     interfaces.PostRepository
	commentRepo  interfaces.CommentRepository
	likeRepo     interfaces.LikeRepository
	userRepo     interfaces.UserRepository
	friendshipRepo interfaces.FriendshipRepository
	strategy     interfaces.FeedSortStrategy
}

// NewFeedService creates a new feed service
func NewFeedService(
	postRepo interfaces.PostRepository,
	commentRepo interfaces.CommentRepository,
	likeRepo interfaces.LikeRepository,
	userRepo interfaces.UserRepository,
	friendshipRepo interfaces.FriendshipRepository,
	strategy interfaces.FeedSortStrategy,
) *FeedService {
	return &FeedService{
		postRepo:       postRepo,
		commentRepo:    commentRepo,
		likeRepo:       likeRepo,
		userRepo:       userRepo,
		friendshipRepo: friendshipRepo,
		strategy:       strategy,
	}
}

// SetStrategy allows changing the feed sort strategy at runtime (Strategy pattern)
func (s *FeedService) SetStrategy(strategy interfaces.FeedSortStrategy) {
	s.strategy = strategy
}

// GetFeed returns the news feed for a user (posts from friends/followed users)
// Algorithm:
// 1. Get all friends/following for the user
// 2. Include the user's own posts
// 3. Fetch posts from all these authors
// 4. Build FeedItems with engagement metrics
// 5. Apply sort strategy (chronological or popularity)
// 6. Paginate results
func (s *FeedService) GetFeed(userID string, limit, offset int) ([]*interfaces.FeedItem, error) {
	// Get friends/following - users whose posts we want to show
	friends, err := s.friendshipRepo.GetAcceptedFriends(userID)
	if err != nil {
		return nil, err
	}

	// Include self in feed authors
	authorIDs := make(map[string]bool)
	authorIDs[userID] = true
	for _, f := range friends {
		authorIDs[f] = true
	}

	authorIDList := make([]string, 0, len(authorIDs))
	for id := range authorIDs {
		authorIDList = append(authorIDList, id)
	}

	// Fetch more posts than needed for sorting (we'll paginate after sorting)
	fetchLimit := limit + offset + 100
	posts, err := s.postRepo.GetPostsByAuthors(authorIDList, fetchLimit, 0)
	if err != nil {
		return nil, err
	}

	// Build feed items with engagement metrics
	items := make([]*interfaces.FeedItem, 0, len(posts))
	authorNames := make(map[string]string)

	for _, post := range posts {
		authorName := authorNames[post.AuthorID]
		if authorName == "" {
			if user, err := s.userRepo.GetByID(post.AuthorID); err == nil {
				authorName = user.Username
				authorNames[post.AuthorID] = authorName
			}
		}

		likeCount := s.likeRepo.GetPostLikeCount(post.ID)
		comments, _ := s.commentRepo.GetByPostID(post.ID)
		commentCount := len(comments)

		items = append(items, &interfaces.FeedItem{
			Post:         post,
			LikeCount:    likeCount,
			CommentCount: commentCount,
			AuthorName:   authorName,
		})
	}

	// Apply strategy (chronological or popularity)
	sorted := s.strategy.Sort(items)

	// Paginate
	start := offset
	if start > len(sorted) {
		return []*interfaces.FeedItem{}, nil
	}
	end := start + limit
	if end > len(sorted) {
		end = len(sorted)
	}
	return sorted[start:end], nil
}
