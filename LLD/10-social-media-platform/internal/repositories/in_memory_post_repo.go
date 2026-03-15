package repositories

import (
	"errors"
	"sort"
	"sync"

	"social-media-platform/internal/models"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
)

// InMemoryPostRepository implements PostRepository with thread-safe in-memory storage.
type InMemoryPostRepository struct {
	posts    map[string]*models.Post
	byAuthor map[string][]string
	mu       sync.RWMutex
}

// InMemoryCommentRepository implements CommentRepository.
type InMemoryCommentRepository struct {
	comments   map[string]*models.Comment
	byPostID   map[string][]string
	mu         sync.RWMutex
}

// InMemoryLikeRepository implements LikeRepository.
type InMemoryLikeRepository struct {
	postLikes    map[string]map[string]bool    // postID -> userID -> true
	commentLikes map[string]map[string]bool    // commentID -> userID -> true
	mu           sync.RWMutex
}

// NewInMemoryPostRepository creates a new in-memory post repository
func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts:    make(map[string]*models.Post),
		byAuthor: make(map[string][]string),
	}
}

// NewInMemoryCommentRepository creates a new in-memory comment repository
func NewInMemoryCommentRepository() *InMemoryCommentRepository {
	return &InMemoryCommentRepository{
		comments: make(map[string]*models.Comment),
		byPostID: make(map[string][]string),
	}
}

// NewInMemoryLikeRepository creates a new in-memory like repository
func NewInMemoryLikeRepository() *InMemoryLikeRepository {
	return &InMemoryLikeRepository{
		postLikes:    make(map[string]map[string]bool),
		commentLikes: make(map[string]map[string]bool),
	}
}

// PostRepository implementation
func (r *InMemoryPostRepository) CreatePost(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts[post.ID] = post
	r.byAuthor[post.AuthorID] = append(r.byAuthor[post.AuthorID], post.ID)
	return nil
}

func (r *InMemoryPostRepository) GetPost(id string) (*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, ErrPostNotFound
	}
	return post, nil
}

func (r *InMemoryPostRepository) UpdatePost(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.ID]; !exists {
		return ErrPostNotFound
	}
	r.posts[post.ID] = post
	return nil
}

func (r *InMemoryPostRepository) DeletePost(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[id]
	if !exists {
		return ErrPostNotFound
	}

	delete(r.posts, id)
	postIDs := r.byAuthor[post.AuthorID]
	for i, pid := range postIDs {
		if pid == id {
			r.byAuthor[post.AuthorID] = append(postIDs[:i], postIDs[i+1:]...)
			break
		}
	}
	return nil
}

func (r *InMemoryPostRepository) GetPostsByAuthors(authorIDs []string, limit, offset int) ([]*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	authorSet := make(map[string]bool)
	for _, id := range authorIDs {
		authorSet[id] = true
	}

	var allPosts []*models.Post
	for _, post := range r.posts {
		if authorSet[post.AuthorID] {
			allPosts = append(allPosts, post)
		}
	}
	sort.Slice(allPosts, func(i, j int) bool {
		if !allPosts[i].CreatedAt.Equal(allPosts[j].CreatedAt) {
			return allPosts[i].CreatedAt.After(allPosts[j].CreatedAt)
		}
		return allPosts[i].ID > allPosts[j].ID // deterministic: newer IDs first when timestamps equal
	})

	start := offset
	if start > len(allPosts) {
		return []*models.Post{}, nil
	}
	end := start + limit
	if end > len(allPosts) {
		end = len(allPosts)
	}
	return allPosts[start:end], nil
}

// CommentRepository implementation
func (r *InMemoryCommentRepository) Create(comment *models.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.comments[comment.ID] = comment
	r.byPostID[comment.PostID] = append(r.byPostID[comment.PostID], comment.ID)
	return nil
}

func (r *InMemoryCommentRepository) GetByID(id string) (*models.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, exists := r.comments[id]
	if !exists {
		return nil, ErrCommentNotFound
	}
	return c, nil
}

func (r *InMemoryCommentRepository) GetByPostID(postID string) ([]*models.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commentIDs, exists := r.byPostID[postID]
	if !exists {
		return []*models.Comment{}, nil
	}

	var comments []*models.Comment
	for _, cid := range commentIDs {
		if c, ok := r.comments[cid]; ok {
			comments = append(comments, c)
		}
	}
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})
	return comments, nil
}

func (r *InMemoryCommentRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	comment, exists := r.comments[id]
	if !exists {
		return ErrCommentNotFound
	}

	delete(r.comments, id)
	commentIDs := r.byPostID[comment.PostID]
	for i, cid := range commentIDs {
		if cid == id {
			r.byPostID[comment.PostID] = append(commentIDs[:i], commentIDs[i+1:]...)
			break
		}
	}
	return nil
}

// LikeRepository implementation
func (r *InMemoryLikeRepository) AddPostLike(postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.postLikes[postID] == nil {
		r.postLikes[postID] = make(map[string]bool)
	}
	r.postLikes[postID][userID] = true
	return nil
}

func (r *InMemoryLikeRepository) RemovePostLike(postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.postLikes[postID] != nil {
		delete(r.postLikes[postID], userID)
	}
	return nil
}

func (r *InMemoryLikeRepository) AddCommentLike(commentID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.commentLikes[commentID] == nil {
		r.commentLikes[commentID] = make(map[string]bool)
	}
	r.commentLikes[commentID][userID] = true
	return nil
}

func (r *InMemoryLikeRepository) RemoveCommentLike(commentID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.commentLikes[commentID] != nil {
		delete(r.commentLikes[commentID], userID)
	}
	return nil
}

func (r *InMemoryLikeRepository) GetPostLikeCount(postID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.postLikes[postID])
}

func (r *InMemoryLikeRepository) GetCommentLikeCount(commentID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.commentLikes[commentID])
}

func (r *InMemoryLikeRepository) HasUserLikedPost(postID, userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.postLikes[postID] != nil && r.postLikes[postID][userID]
}

func (r *InMemoryLikeRepository) HasUserLikedComment(commentID, userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.commentLikes[commentID] != nil && r.commentLikes[commentID][userID]
}
