package tests

import (
	"chat-application/internal/broker"
	"chat-application/internal/models"
	"chat-application/internal/repositories"
	"chat-application/internal/services"
	"testing"
)

func setupChatTest(t *testing.T) (*services.ChatService, *services.AuthService, string, string) {
	userRepo := repositories.NewInMemoryUserRepository()
	roomRepo := repositories.NewInMemoryChatRoomRepository()
	authService := services.NewAuthService(userRepo)
	chatService := services.NewChatService(roomRepo, userRepo)

	user1, _ := authService.Register("alice", "alice@test.com", "pass")
	user2, _ := authService.Register("bob", "bob@test.com", "pass")

	return chatService, authService, user1.ID, user2.ID
}

func TestChatService_CreateOneOnOneRoom(t *testing.T) {
	chatService, _, user1ID, user2ID := setupChatTest(t)

	room, err := chatService.CreateOneOnOneRoom(user1ID, user2ID)
	if err != nil {
		t.Fatalf("CreateOneOnOneRoom failed: %v", err)
	}
	if room.Type != models.ChatRoomTypeOneOnOne {
		t.Errorf("expected one_on_one, got %s", room.Type)
	}
	if len(room.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(room.Members))
	}
	if !room.HasMember(user1ID) || !room.HasMember(user2ID) {
		t.Error("expected both users to be members")
	}
}

func TestChatService_CreateOneOnOneRoomIdempotent(t *testing.T) {
	chatService, _, user1ID, user2ID := setupChatTest(t)

	room1, _ := chatService.CreateOneOnOneRoom(user1ID, user2ID)
	room2, _ := chatService.CreateOneOnOneRoom(user2ID, user1ID)

	if room1.ID != room2.ID {
		t.Errorf("same users should get same room: %s vs %s", room1.ID, room2.ID)
	}
}

func TestChatService_CreateGroupRoom(t *testing.T) {
	chatService, _, user1ID, user2ID := setupChatTest(t)

	room, err := chatService.CreateGroupRoom(user1ID, "Team", []string{user2ID})
	if err != nil {
		t.Fatalf("CreateGroupRoom failed: %v", err)
	}
	if room.Type != models.ChatRoomTypeGroup {
		t.Errorf("expected group, got %s", room.Type)
	}
	if room.Name != "Team" {
		t.Errorf("expected name Team, got %s", room.Name)
	}
	if len(room.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(room.Members))
	}
}

func TestChatService_AddMember(t *testing.T) {
	chatService, authService, user1ID, user2ID := setupChatTest(t)
	user3, _ := authService.Register("charlie", "charlie@test.com", "pass")

	room, _ := chatService.CreateGroupRoom(user1ID, "Team", []string{user2ID})

	err := chatService.AddMember(room.ID, user1ID, user3.ID)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	updated, _ := chatService.GetRoom(room.ID)
	if !updated.HasMember(user3.ID) {
		t.Error("expected charlie to be added")
	}
}

func TestChatService_AddMemberNotAdmin(t *testing.T) {
	chatService, authService, user1ID, user2ID := setupChatTest(t)
	user3, _ := authService.Register("charlie", "charlie@test.com", "pass")

	room, _ := chatService.CreateGroupRoom(user1ID, "Team", []string{user2ID})

	err := chatService.AddMember(room.ID, user2ID, user3.ID)
	if err != services.ErrCannotManage {
		t.Errorf("expected ErrCannotManage, got %v", err)
	}
}

func TestChatService_LeaveRoom(t *testing.T) {
	chatService, _, user1ID, user2ID := setupChatTest(t)

	room, _ := chatService.CreateGroupRoom(user1ID, "Team", []string{user2ID})

	err := chatService.LeaveRoom(room.ID, user2ID)
	if err != nil {
		t.Fatalf("LeaveRoom failed: %v", err)
	}

	updated, _ := chatService.GetRoom(room.ID)
	if updated.HasMember(user2ID) {
		t.Error("bob should have left the room")
	}
}

func TestChatService_GetUserRooms(t *testing.T) {
	chatService, _, user1ID, user2ID := setupChatTest(t)

	_, _ = chatService.CreateOneOnOneRoom(user1ID, user2ID)
	_, _ = chatService.CreateGroupRoom(user1ID, "Group", []string{user2ID})

	rooms, err := chatService.GetUserRooms(user1ID)
	if err != nil {
		t.Fatalf("GetUserRooms failed: %v", err)
	}
	if len(rooms) < 2 {
		t.Errorf("expected at least 2 rooms, got %d", len(rooms))
	}
}

// Ensure broker is used (compile check)
var _ = broker.NewInMemoryBroker(&broker.DirectDelivery{})
