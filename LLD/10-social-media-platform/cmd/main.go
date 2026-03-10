package main

import (
	"fmt"
	"log"

	"social-media-platform/internal/repositories"
	"social-media-platform/internal/services"
)

func main() {
	// Initialize repositories (Repository pattern)
	userRepo := repositories.NewInMemoryUserRepository()
	friendshipRepo := repositories.NewInMemoryFriendshipRepository()
	postRepo := repositories.NewInMemoryPostRepository()
	commentRepo := repositories.NewInMemoryCommentRepository()
	likeRepo := repositories.NewInMemoryLikeRepository()
	notifRepo := repositories.NewInMemoryNotificationRepository()

	// Initialize notification publisher (Observer pattern)
	publisher := services.NewNotificationPublisher()
	notifService := services.NewNotificationService(notifRepo, publisher)

	// Initialize services
	userService := services.NewUserService(userRepo)
	friendshipService := services.NewFriendshipService(userRepo, friendshipRepo, notifService, publisher)
	postService := services.NewPostService(postRepo, commentRepo, likeRepo, userRepo, notifService)

	// Feed with chronological strategy (Strategy pattern)
	chronologicalStrategy := &services.ChronologicalFeedStrategy{}
	feedService := services.NewFeedService(
		postRepo, commentRepo, likeRepo, userRepo, friendshipRepo,
		chronologicalStrategy,
	)

	// Demo: Register users
	alice, err := userService.Register("alice", "alice@example.com", "Hello world", "")
	if err != nil {
		log.Fatal(err)
	}
	bob, err := userService.Register("bob", "bob@example.com", "Bob's bio", "")
	if err != nil {
		log.Fatal(err)
	}
	charlie, err := userService.Register("charlie", "charlie@example.com", "Charlie here", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Registered users: %s, %s, %s\n", alice.Username, bob.Username, charlie.Username)

	// Demo: Friend requests
	fs1, err := friendshipService.SendFriendRequest(alice.ID, bob.ID)
	if err != nil {
		log.Fatal(err)
	}
	_ = friendshipService.AcceptFriendRequest(bob.ID, fs1.ID)

	fs2, _ := friendshipService.SendFriendRequest(bob.ID, charlie.ID)
	_ = friendshipService.AcceptFriendRequest(charlie.ID, fs2.ID)

	fmt.Println("Friendships established: Alice-Bob, Bob-Charlie")

	// Demo: Create posts
	post1, _ := postService.CreatePost(alice.ID, "Hello from Alice!", []string{})
	post2, _ := postService.CreatePost(bob.ID, "Bob's first post", []string{"https://example.com/img1.jpg"})
	post3, _ := postService.CreatePost(charlie.ID, "Charlie says hi", nil)

	fmt.Printf("Created posts: %s, %s, %s\n", post1.ID, post2.ID, post3.ID)

	// Demo: Comments and likes
	_, _ = postService.AddComment(post1.ID, bob.ID, "Nice post Alice!")
	_ = postService.LikePost(post1.ID, bob.ID)
	_ = postService.LikePost(post1.ID, charlie.ID)
	_, _ = postService.AddComment(post2.ID, alice.ID, "Great photo Bob!")

	fmt.Println("Added comments and likes")

	// Demo: Feed for Bob (friends: Alice, Charlie)
	feed, err := feedService.GetFeed(bob.ID, 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Feed for %s (chronological) ===\n", bob.Username)
	for i, item := range feed {
		fmt.Printf("%d. [%s] %s (likes: %d, comments: %d)\n",
			i+1, item.AuthorName, item.Post.Content, item.LikeCount, item.CommentCount)
	}

	// Switch to popularity strategy
	feedService.SetStrategy(&services.PopularityFeedStrategy{})
	feed, _ = feedService.GetFeed(bob.ID, 10, 0)
	fmt.Printf("\n=== Feed for %s (popularity) ===\n", bob.Username)
	for i, item := range feed {
		fmt.Printf("%d. [%s] %s (likes: %d, comments: %d)\n",
			i+1, item.AuthorName, item.Post.Content, item.LikeCount, item.CommentCount)
	}

	// Demo: Notifications for Alice (received like and comment)
	notifs, _ := notifService.GetUserNotifications(alice.ID, 10, 0)
	fmt.Printf("\n=== Notifications for %s ===\n", alice.Username)
	for _, n := range notifs {
		fmt.Printf("- %s from user %s\n", n.Type, n.SourceUserID)
	}

	// Demo: User search
	results, _ := userService.SearchUsers("alice", 5)
	fmt.Printf("\n=== Search 'alice': %d results ===\n", len(results))
	for _, u := range results {
		fmt.Printf("- %s (%s)\n", u.Username, u.Email)
	}

	fmt.Println("\nSocial Media Platform demo completed successfully!")
}
