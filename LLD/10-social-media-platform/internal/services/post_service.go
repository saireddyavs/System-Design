package services

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"social-media-platform/internal/interfaces"
	"social-media-platform/internal/models"
)

var (
	ErrInvalidPost       = errors.New("invalid post content")
	ErrInvalidComment    = errors.New("invalid comment content")
	ErrPostNotFound      = errors.New("post not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrUnauthorizedEdit  = errors.New("unauthorized to edit this resource")
)

// PostService handles post, comment, and like operations.
// Observer: Triggers notifications on like/comment events
type PostService struct {
	postRepo     interfaces.PostRepository
	commentRepo  interfaces.CommentRepository
	likeRepo     interfaces.LikeRepository
	userRepo     interfaces.UserRepository
	notifService *NotificationService
	nextPostID   atomic.Uint64
	nextCommentID atomic.Uint64
	mu           sync.Mutex
}

// NewPostService creates a new post service
func NewPostService(
	postRepo interfaces.PostRepository,
	commentRepo interfaces.CommentRepository,
	likeRepo interfaces.LikeRepository,
	userRepo interfaces.UserRepository,
	notifService *NotificationService,
) *PostService {
	return &PostService{
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		likeRepo:     likeRepo,
		userRepo:     userRepo,
		notifService: notifService,
	}
}

// CreatePost creates a new post (Factory: uses models.NewPost)
func (s *PostService) CreatePost(authorID, content string, imageURLs []string) (*models.Post, error) {
	if content == "" && len(imageURLs) == 0 {
		return nil, ErrInvalidPost
	}

	if _, err := s.userRepo.GetByID(authorID); err != nil {
		return nil, err
	}

	s.mu.Lock()
	id := fmt.Sprintf("post-%d", s.nextPostID.Add(1))
	s.mu.Unlock()

	post := models.NewPost(id, authorID, content, imageURLs)
	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, err
	}
	return post, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(postID string) (*models.Post, error) {
	return s.postRepo.GetPost(postID)
}

// UpdatePost updates a post (author only)
func (s *PostService) UpdatePost(postID, userID, content string, imageURLs []string) (*models.Post, error) {
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		return nil, err
	}
	if post.AuthorID != userID {
		return nil, ErrUnauthorizedEdit
	}

	post.Update(content, imageURLs)
	return post, s.postRepo.UpdatePost(post)
}

// DeletePost deletes a post (author only)
func (s *PostService) DeletePost(postID, userID string) error {
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		return err
	}
	if post.AuthorID != userID {
		return ErrUnauthorizedEdit
	}
	return s.postRepo.DeletePost(postID)
}

// AddComment adds a comment to a post (Observer: triggers notification)
func (s *PostService) AddComment(postID, authorID, content string) (*models.Comment, error) {
	if content == "" {
		return nil, ErrInvalidComment
	}

	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	id := fmt.Sprintf("comment-%d", s.nextCommentID.Add(1))
	s.mu.Unlock()

	comment := models.NewComment(id, postID, authorID, content)
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	// Observer: Notify post author (if not self-comment)
	if post.AuthorID != authorID {
		notif := models.NewNotification(
			s.notifService.GenerateID(),
			post.AuthorID,
			models.NotificationTypeComment,
			authorID,
			comment.ID,
		)
		_ = s.notifService.CreateAndPublish(notif)
	}

	return comment, nil
}

// DeleteComment deletes a comment (author only)
func (s *PostService) DeleteComment(commentID, userID string) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	if comment.AuthorID != userID {
		return ErrUnauthorizedEdit
	}
	return s.commentRepo.Delete(commentID)
}

// LikePost adds a like to a post (Observer: triggers notification)
func (s *PostService) LikePost(postID, userID string) error {
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		return err
	}
	if s.likeRepo.HasUserLikedPost(postID, userID) {
		return nil // Already liked
	}

	if err := s.likeRepo.AddPostLike(postID, userID); err != nil {
		return err
	}

	// Observer: Notify post author (if not self-like)
	if post.AuthorID != userID {
		notif := models.NewNotification(
			s.notifService.GenerateID(),
			post.AuthorID,
			models.NotificationTypeLike,
			userID,
			postID,
		)
		_ = s.notifService.CreateAndPublish(notif)
	}

	return nil
}

// UnlikePost removes a like from a post
func (s *PostService) UnlikePost(postID, userID string) error {
	return s.likeRepo.RemovePostLike(postID, userID)
}

// LikeComment adds a like to a comment
func (s *PostService) LikeComment(commentID, userID string) error {
	if s.likeRepo.HasUserLikedComment(commentID, userID) {
		return nil
	}
	return s.likeRepo.AddCommentLike(commentID, userID)
}

// UnlikeComment removes a like from a comment
func (s *PostService) UnlikeComment(commentID, userID string) error {
	return s.likeRepo.RemoveCommentLike(commentID, userID)
}

// GetPostWithDetails returns post with like count and comments
func (s *PostService) GetPostWithDetails(postID string) (*models.Post, int, []*models.Comment, error) {
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		return nil, 0, nil, err
	}
	likeCount := s.likeRepo.GetPostLikeCount(postID)
	comments, err := s.commentRepo.GetByPostID(postID)
	if err != nil {
		return nil, 0, nil, err
	}
	return post, likeCount, comments, nil
}

// GetComments returns comments for a post
func (s *PostService) GetComments(postID string) ([]*models.Comment, error) {
	return s.commentRepo.GetByPostID(postID)
}
