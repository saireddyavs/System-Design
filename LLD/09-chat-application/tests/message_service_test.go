package tests

import (
	"chat-application/internal/broker"
	"chat-application/internal/models"
	"chat-application/internal/repositories"
	"chat-application/internal/services"
	"sync"
	"testing"
	"time"
)

func setupMessageTest(t *testing.T) (*services.MessageService, *services.ChatService, *services.AuthService, string, string, string) {
	userRepo := repositories.NewInMemoryUserRepository()
	msgRepo := repositories.NewInMemoryMessageRepository()
	roomRepo := repositories.NewInMemoryChatRoomRepository()
	msgBroker := broker.NewInMemoryBroker(&broker.DirectDelivery{})

	authService := services.NewAuthService(userRepo)
	chatService := services.NewChatService(roomRepo, userRepo)
	msgFactory := services.NewMessageFactory()
	msgService := services.NewMessageService(msgRepo, roomRepo, msgBroker, msgFactory)

	user1, _ := authService.Register("alice", "alice@test.com", "pass")
	user2, _ := authService.Register("bob", "bob@test.com", "pass")

	room, _ := chatService.CreateOneOnOneRoom(user1.ID, user2.ID)
	return msgService, chatService, authService, user1.ID, user2.ID, room.ID
}

func TestMessageService_SendMessage(t *testing.T) {
	msgService, _, _, user1ID, _, roomID := setupMessageTest(t)

	msg, err := msgService.SendMessage(user1ID, roomID, "Hello!", models.MessageTypeText)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if msg.Content != "Hello!" {
		t.Errorf("expected content Hello!, got %s", msg.Content)
	}
	if msg.SenderID != user1ID {
		t.Errorf("expected sender %s, got %s", user1ID, msg.SenderID)
	}
	if msg.Status != models.MessageStatusSent {
		t.Errorf("expected status sent, got %s", msg.Status)
	}
}

func TestMessageService_SendMessageNotInRoom(t *testing.T) {
	msgService, _, authService, _, _, roomID := setupMessageTest(t)
	user3, _ := authService.Register("charlie", "charlie@test.com", "pass")

	_, err := msgService.SendMessage(user3.ID, roomID, "Hello!", models.MessageTypeText)
	if err != services.ErrNotInRoom {
		t.Errorf("expected ErrNotInRoom, got %v", err)
	}
}

func TestMessageService_GetMessageHistory(t *testing.T) {
	msgService, _, _, user1ID, user2ID, roomID := setupMessageTest(t)

	msgService.SendMessage(user1ID, roomID, "Msg1", models.MessageTypeText)
	msgService.SendMessage(user2ID, roomID, "Msg2", models.MessageTypeText)

	history, err := msgService.GetMessageHistory(roomID, user1ID, 10, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 messages, got %d", len(history))
	}
}

func TestMessageService_GetMessageHistoryPagination(t *testing.T) {
	msgService, _, _, user1ID, _, roomID := setupMessageTest(t)

	for i := 0; i < 5; i++ {
		msgService.SendMessage(user1ID, roomID, "Msg", models.MessageTypeText)
	}

	history, _ := msgService.GetMessageHistory(roomID, user1ID, 2, 1)
	if len(history) != 2 {
		t.Errorf("expected 2 messages with limit 2 offset 1, got %d", len(history))
	}
}

func TestMessageService_MarkAsRead(t *testing.T) {
	msgService, _, _, user1ID, user2ID, roomID := setupMessageTest(t)

	msg, _ := msgService.SendMessage(user1ID, roomID, "Hello", models.MessageTypeText)

	err := msgService.MarkAsRead(msg.ID, user2ID)
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	// Verify via history - message should show read
	history, _ := msgService.GetMessageHistory(roomID, user2ID, 10, 0)
	found := false
	for _, m := range history {
		if m.ID == msg.ID && m.IsReadBy(user2ID) {
			found = true
			break
		}
	}
	if !found {
		t.Error("message should be marked as read by bob")
	}
}

func TestMessageService_RealTimeDelivery(t *testing.T) {
	msgService, _, _, user1ID, user2ID, roomID := setupMessageTest(t)

	ch := msgService.Subscribe(user2ID)
	defer msgService.Unsubscribe(user2ID)

	var received *models.Message
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case received = <-ch:
			return
		case <-time.After(2 * time.Second):
			return
		}
	}()

	time.Sleep(50 * time.Millisecond)
	msgService.SendMessage(user1ID, roomID, "Real-time test", models.MessageTypeText)

	wg.Wait()
	if received == nil {
		t.Fatal("expected to receive message via real-time channel")
	}
	if received.Content != "Real-time test" {
		t.Errorf("expected content 'Real-time test', got %s", received.Content)
	}
}

func TestMessageService_ConcurrentSends(t *testing.T) {
	msgService, _, _, user1ID, user2ID, roomID := setupMessageTest(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = msgService.SendMessage(user1ID, roomID, "Concurrent", models.MessageTypeText)
			_, _ = msgService.SendMessage(user2ID, roomID, "Concurrent", models.MessageTypeText)
		}(i)
	}
	wg.Wait()

	history, _ := msgService.GetMessageHistory(roomID, user1ID, 100, 0)
	if len(history) != 20 {
		t.Errorf("expected 20 messages from concurrent sends, got %d", len(history))
	}
}
