package interfaces

import (
	"social-media-platform/internal/models"
)

// FeedItem represents a post with engagement metrics for feed display
type FeedItem struct {
	Post       *models.Post
	LikeCount  int
	CommentCount int
	AuthorName string
}

// FeedSortStrategy defines the strategy for sorting feed items (Strategy pattern).
// O - Open/Closed: New sorting algorithms can be added without modifying existing code
type FeedSortStrategy interface {
	Sort(items []*FeedItem) []*FeedItem
	Name() string
}
