package tests

import (
	"testing"

	"social-media-platform/internal/models"
	"social-media-platform/internal/repositories"
	"social-media-platform/internal/services"
)

func setupFriendshipService(t *testing.T) (*services.FriendshipService, *services.UserService, *services.NotificationService) {
	userRepo := repositories.NewInMemoryUserRepository()
	friendshipRepo := repositories.NewInMemoryFriendshipRepository()
	notifRepo := repositories.NewInMemoryNotificationRepository()
	publisher := services.NewNotificationPublisher()
	notifService := services.NewNotificationService(notifRepo, publisher)

	userService := services.NewUserService(userRepo)
	friendshipService := services.NewFriendshipService(userRepo, friendshipRepo, notifService, publisher)
	return friendshipService, userService, notifService
}

func TestSendFriendRequest(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	friendship, err := fs.SendFriendRequest(alice.ID, bob.ID)
	if err != nil {
		t.Fatalf("SendFriendRequest failed: %v", err)
	}
	if friendship.Status != models.FriendshipStatusPending {
		t.Errorf("expected status Pending, got %s", friendship.Status)
	}
	if friendship.RequesterID != alice.ID {
		t.Errorf("expected requester %s, got %s", alice.ID, friendship.RequesterID)
	}
	if friendship.ReceiverID != bob.ID {
		t.Errorf("expected receiver %s, got %s", bob.ID, friendship.ReceiverID)
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	fs, userService, notifService := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	friendship, _ := fs.SendFriendRequest(alice.ID, bob.ID)

	err := fs.AcceptFriendRequest(bob.ID, friendship.ID)
	if err != nil {
		t.Fatalf("AcceptFriendRequest failed: %v", err)
	}

	friends, _ := fs.GetFriends(bob.ID)
	if len(friends) != 1 {
		t.Errorf("expected 1 friend, got %d", len(friends))
	}
	if friends[0] != alice.ID {
		t.Errorf("expected friend alice, got %s", friends[0])
	}

	// Verify notification to requester
	notifs, _ := notifService.GetUserNotifications(alice.ID, 10, 0)
	found := false
	for _, n := range notifs {
		if n.Type == models.NotificationTypeFriendAccepted {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected friend_accepted notification for requester")
	}
}

func TestRejectFriendRequest(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	friendship, _ := fs.SendFriendRequest(alice.ID, bob.ID)
	err := fs.RejectFriendRequest(bob.ID, friendship.ID)
	if err != nil {
		t.Fatalf("RejectFriendRequest failed: %v", err)
	}

	friends, _ := fs.GetFriends(bob.ID)
	if len(friends) != 0 {
		t.Errorf("expected 0 friends after reject, got %d", len(friends))
	}
}

func TestUnfollow(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	friendship, _ := fs.SendFriendRequest(alice.ID, bob.ID)
	_ = fs.AcceptFriendRequest(bob.ID, friendship.ID)

	err := fs.Unfollow(alice.ID, bob.ID)
	if err != nil {
		t.Fatalf("Unfollow failed: %v", err)
	}

	friends, _ := fs.GetFriends(bob.ID)
	if len(friends) != 0 {
		t.Errorf("expected 0 friends after unfollow, got %d", len(friends))
	}
}

func TestCannotFriendSelf(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")

	_, err := fs.SendFriendRequest(alice.ID, alice.ID)
	if err == nil {
		t.Error("expected error when sending friend request to self")
	}
}

func TestDuplicateFriendRequest(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	_, _ = fs.SendFriendRequest(alice.ID, bob.ID)
	_, err := fs.SendFriendRequest(alice.ID, bob.ID)
	if err == nil {
		t.Error("expected error for duplicate friend request")
	}
}

func TestFriendRequestNotification(t *testing.T) {
	fs, userService, notifService := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")

	_, _ = fs.SendFriendRequest(alice.ID, bob.ID)

	notifs, _ := notifService.GetUserNotifications(bob.ID, 10, 0)
	if len(notifs) == 0 {
		t.Error("expected notification for receiver on friend request")
	}
	found := false
	for _, n := range notifs {
		if n.Type == models.NotificationTypeFriendRequest {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected friend_request notification for receiver")
	}
}

func TestGetPendingRequests(t *testing.T) {
	fs, userService, _ := setupFriendshipService(t)

	alice, _ := userService.Register("alice", "alice@test.com", "", "")
	bob, _ := userService.Register("bob", "bob@test.com", "", "")
	charlie, _ := userService.Register("charlie", "charlie@test.com", "", "")

	_, _ = fs.SendFriendRequest(alice.ID, bob.ID)
	_, _ = fs.SendFriendRequest(charlie.ID, bob.ID)

	pending, _ := fs.GetPendingRequests(bob.ID)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending requests, got %d", len(pending))
	}
}
