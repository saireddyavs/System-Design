package models

import (
	"sync"
	"time"
)

// Comment represents a comment on a post.
// S - Single Responsibility: Comment model only holds comment data
type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	mu        sync.RWMutex
}

// NewComment creates a new Comment instance (Factory pattern)
func NewComment(id, postID, authorID, content string) *Comment {
	return &Comment{
		ID:        id,
		PostID:    postID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}
