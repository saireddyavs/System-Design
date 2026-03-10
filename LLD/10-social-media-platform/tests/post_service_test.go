package tests

import (
	"testing"

	"social-media-platform/internal/repositories"
	"social-media-platform/internal/services"
)

func setupPostService(t *testing.T) (*services.PostService, *services.NotificationService, *repositories.InMemoryUserRepository) {
	userRepo := repositories.NewInMemoryUserRepository()
	postRepo := repositories.NewInMemoryPostRepository()
	commentRepo := repositories.NewInMemoryCommentRepository()
	likeRepo := repositories.NewInMemoryLikeRepository()
	notifRepo := repositories.NewInMemoryNotificationRepository()
	publisher := services.NewNotificationPublisher()
	notifService := services.NewNotificationService(notifRepo, publisher)

	postService := services.NewPostService(
		postRepo, commentRepo, likeRepo, userRepo, notifService,
	)
	return postService, notifService, userRepo
}

func TestCreatePost(t *testing.T) {
	postService, _, userRepo := setupPostService(t)

	userSvc := services.NewUserService(userRepo)
	u, err := userSvc.Register("testuser", "test@test.com", "bio", "")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	post, err := postService.CreatePost(u.ID, "Hello World", nil)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	if post == nil {
		t.Fatal("expected post, got nil")
	}
	if post.Content != "Hello World" {
		t.Errorf("expected content 'Hello World', got %s", post.Content)
	}
	if post.AuthorID != u.ID {
		t.Errorf("expected author %s, got %s", u.ID, post.AuthorID)
	}
}

func TestCreatePostEmptyContent(t *testing.T) {
	postService, _, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u, _ := userSvc.Register("testuser2", "test2@test.com", "", "")

	_, err := postService.CreatePost(u.ID, "", nil)
	if err == nil {
		t.Error("expected error for empty content, got nil")
	}
}

func TestAddComment(t *testing.T) {
	postService, notifService, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u1, _ := userSvc.Register("author", "author@test.com", "", "")
	u2, _ := userSvc.Register("commenter", "commenter@test.com", "", "")

	post, _ := postService.CreatePost(u1.ID, "Post content", nil)

	comment, err := postService.AddComment(post.ID, u2.ID, "Great post!")
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
	if comment.Content != "Great post!" {
		t.Errorf("expected 'Great post!', got %s", comment.Content)
	}
	if comment.PostID != post.ID {
		t.Errorf("expected PostID %s, got %s", post.ID, comment.PostID)
	}

	// Verify notification was created for author
	notifs, _ := notifService.GetUserNotifications(u1.ID, 10, 0)
	if len(notifs) == 0 {
		t.Error("expected notification for post author, got none")
	}
}

func TestLikePost(t *testing.T) {
	postService, notifService, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u1, _ := userSvc.Register("author", "author@test.com", "", "")
	u2, _ := userSvc.Register("liker", "liker@test.com", "", "")

	post, _ := postService.CreatePost(u1.ID, "Post", nil)

	err := postService.LikePost(post.ID, u2.ID)
	if err != nil {
		t.Fatalf("LikePost failed: %v", err)
	}

	_, likeCount, _, _ := postService.GetPostWithDetails(post.ID)
	if likeCount != 1 {
		t.Errorf("expected like count 1, got %d", likeCount)
	}

	// Verify notification
	notifs, _ := notifService.GetUserNotifications(u1.ID, 10, 0)
	foundLike := false
	for _, n := range notifs {
		if n.Type == "like" {
			foundLike = true
			break
		}
	}
	if !foundLike {
		t.Error("expected like notification for author")
	}
}

func TestUnlikePost(t *testing.T) {
	postService, _, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u1, _ := userSvc.Register("author", "a@test.com", "", "")
	u2, _ := userSvc.Register("liker", "l@test.com", "", "")

	post, _ := postService.CreatePost(u1.ID, "Post", nil)
	_ = postService.LikePost(post.ID, u2.ID)
	_ = postService.UnlikePost(post.ID, u2.ID)

	_, likeCount, _, _ := postService.GetPostWithDetails(post.ID)
	if likeCount != 0 {
		t.Errorf("expected like count 0 after unlike, got %d", likeCount)
	}
}

func TestUpdatePost(t *testing.T) {
	postService, _, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u, _ := userSvc.Register("author", "a@test.com", "", "")

	post, _ := postService.CreatePost(u.ID, "Original", nil)
	updated, err := postService.UpdatePost(post.ID, u.ID, "Updated content", nil)
	if err != nil {
		t.Fatalf("UpdatePost failed: %v", err)
	}
	if updated.Content != "Updated content" {
		t.Errorf("expected 'Updated content', got %s", updated.Content)
	}
}

func TestUpdatePostUnauthorized(t *testing.T) {
	postService, _, userRepo := setupPostService(t)
	userSvc := services.NewUserService(userRepo)
	u1, _ := userSvc.Register("author", "a@test.com", "", "")
	u2, _ := userSvc.Register("other", "o@test.com", "", "")

	post, _ := postService.CreatePost(u1.ID, "Post", nil)
	_, err := postService.UpdatePost(post.ID, u2.ID, "Hacked", nil)
	if err == nil {
		t.Error("expected error for unauthorized update")
	}
}
