package tests

import (
	"testing"
	"time"

	"social-media-platform/internal/repositories"
	"social-media-platform/internal/services"
)

func setupFeedService(t *testing.T) (*services.FeedService, *services.UserService, *services.FriendshipService, *services.PostService) {
	userRepo := repositories.NewInMemoryUserRepository()
	friendshipRepo := repositories.NewInMemoryFriendshipRepository()
	postRepo := repositories.NewInMemoryPostRepository()
	commentRepo := repositories.NewInMemoryCommentRepository()
	likeRepo := repositories.NewInMemoryLikeRepository()
	notifRepo := repositories.NewInMemoryNotificationRepository()
	publisher := services.NewNotificationPublisher()
	notifService := services.NewNotificationService(notifRepo, publisher)

	userService := services.NewUserService(userRepo)
	friendshipService := services.NewFriendshipService(userRepo, friendshipRepo, notifService, publisher)
	postService := services.NewPostService(postRepo, commentRepo, likeRepo, userRepo, notifService)

	chronologicalStrategy := &services.ChronologicalFeedStrategy{}
	feedService := services.NewFeedService(
		postRepo, commentRepo, likeRepo, userRepo, friendshipRepo,
		chronologicalStrategy,
	)
	return feedService, userService, friendshipService, postService
}

func TestFeedGeneration(t *testing.T) {
	feedService, userService, friendshipService, postService := setupFeedService(t)

	// Create users
	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	// Create friendship
	fs, _ := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs.ID)

	// Create posts
	alicePost, _ := postService.CreatePost(alice.ID, "Alice's post", nil)
	bobPost, _ := postService.CreatePost(bob.ID, "Bob's post", nil)

	// Bob's feed should include both Alice and Bob (friends)
	feed, err := feedService.GetFeed(bob.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetFeed failed: %v", err)
	}
	if len(feed) != 2 {
		t.Errorf("expected 2 feed items, got %d", len(feed))
	}

	// Verify both posts are in feed
	postIDs := make(map[string]bool)
	for _, item := range feed {
		postIDs[item.Post.ID] = true
	}
	if !postIDs[alicePost.ID] {
		t.Error("Alice's post not in Bob's feed")
	}
	if !postIDs[bobPost.ID] {
		t.Error("Bob's post not in Bob's feed")
	}
}

func TestFeedChronologicalOrder(t *testing.T) {
	feedService, userService, friendshipService, postService := setupFeedService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")
	fs, _ := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs.ID)

	post1, _ := postService.CreatePost(alice.ID, "First", nil)
	time.Sleep(2 * time.Millisecond)
	post2, _ := postService.CreatePost(bob.ID, "Second", nil)
	time.Sleep(2 * time.Millisecond)
	post3, _ := postService.CreatePost(alice.ID, "Third", nil)

	feed, _ := feedService.GetFeed(bob.ID, 10, 0)
	if len(feed) < 3 {
		t.Fatalf("expected at least 3 items, got %d", len(feed))
	}
	// Chronological: newest first
	if feed[0].Post.ID != post3.ID {
		t.Errorf("expected newest post first (Third), got %s", feed[0].Post.Content)
	}
	if feed[1].Post.ID != post2.ID {
		t.Errorf("expected second newest (Second), got %s", feed[1].Post.Content)
	}
	if feed[2].Post.ID != post1.ID {
		t.Errorf("expected oldest (First), got %s", feed[2].Post.Content)
	}
}

func TestFeedPopularityStrategy(t *testing.T) {
	feedService, userService, friendshipService, postService := setupFeedService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")
	fs, _ := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs.ID)

	_, _ = postService.CreatePost(alice.ID, "Low engagement", nil)
	post2, _ := postService.CreatePost(bob.ID, "High engagement", nil)

	// Bob likes and comments on post2
	_ = postService.LikePost(post2.ID, alice.ID)
	_, _ = postService.AddComment(post2.ID, alice.ID, "Great!")

	// Switch to popularity strategy
	feedService.SetStrategy(&services.PopularityFeedStrategy{})
	feed, _ := feedService.GetFeed(bob.ID, 10, 0)

	if len(feed) < 2 {
		t.Fatalf("expected 2 items, got %d", len(feed))
	}
	// High engagement post should be first
	if feed[0].Post.ID != post2.ID {
		t.Errorf("expected high engagement post first, got %s (likes: %d)", feed[0].Post.Content, feed[0].LikeCount)
	}
}

func TestFeedPagination(t *testing.T) {
	feedService, userService, friendshipService, postService := setupFeedService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")
	fs, _ := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs.ID)

	// Create 5 posts with delays to ensure distinct timestamps
	for i := 0; i < 5; i++ {
		_, _ = postService.CreatePost(alice.ID, "Post", nil)
		time.Sleep(2 * time.Millisecond)
	}

	// Fetch first 2
	feed1, _ := feedService.GetFeed(bob.ID, 2, 0)
	if len(feed1) != 2 {
		t.Errorf("expected 2 items, got %d", len(feed1))
	}

	// Fetch next 2
	feed2, _ := feedService.GetFeed(bob.ID, 2, 2)
	if len(feed2) != 2 {
		t.Errorf("expected 2 items on page 2, got %d", len(feed2))
	}

	// No overlap
	for _, item1 := range feed1 {
		for _, item2 := range feed2 {
			if item1.Post.ID == item2.Post.ID {
				t.Error("pagination overlap detected")
			}
		}
	}
}

func TestFeedExcludesNonFriends(t *testing.T) {
	feedService, userService, friendshipService, postService := setupFeedService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")
	charlie, _ := userService.Register("charlie", "charlie@test.com", "", "")

	// Alice and Bob are friends, Charlie is not
	fs, _ := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs.ID)

	_, _ = postService.CreatePost(alice.ID, "Alice post", nil)
	charliePost, _ := postService.CreatePost(charlie.ID, "Charlie post", nil)

	// Bob's feed should NOT include Charlie's post
	feed, _ := feedService.GetFeed(bob.ID, 10, 0)
	for _, item := range feed {
		if item.Post.ID == charliePost.ID {
			t.Error("Charlie's post should not appear in Bob's feed (not friends)")
		}
	}
}
