package models

import (
	"sync"
	"time"
)

// Post represents a social media post.
// S - Single Responsibility: Post model only holds post data
type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	ImageURLs []string  `json:"image_urls"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	mu        sync.RWMutex
}

// NewPost creates a new Post instance (Factory pattern)
func NewPost(id, authorID, content string, imageURLs []string) *Post {
	now := time.Now().UTC()
	return &Post{
		ID:        id,
		AuthorID:  authorID,
		Content:   content,
		ImageURLs: imageURLs,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates post content and image URLs
func (p *Post) Update(content string, imageURLs []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	if content != "" {
		p.Content = content
	}
	if imageURLs != nil {
		p.ImageURLs = imageURLs
	}
}
