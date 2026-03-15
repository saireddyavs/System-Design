package interfaces

import (
	"social-media-platform/internal/models"
)

// PostRepository defines the contract for post data access (Repository pattern).
// D - Dependency Inversion: Services depend on this interface
type PostRepository interface {
	CreatePost(post *models.Post) error
	GetPost(id string) (*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id string) error
	GetPostsByAuthors(authorIDs []string, limit, offset int) ([]*models.Post, error)
}

// CommentRepository defines the contract for comment data access.
type CommentRepository interface {
	Create(comment *models.Comment) error
	GetByID(id string) (*models.Comment, error)
	GetByPostID(postID string) ([]*models.Comment, error)
	Delete(id string) error
}

// LikeRepository defines the contract for like data access.
type LikeRepository interface {
	AddPostLike(postID, userID string) error
	RemovePostLike(postID, userID string) error
	AddCommentLike(commentID, userID string) error
	RemoveCommentLike(commentID, userID string) error
	GetPostLikeCount(postID string) int
	GetCommentLikeCount(commentID string) int
	HasUserLikedPost(postID, userID string) bool
	HasUserLikedComment(commentID, userID string) bool
}
