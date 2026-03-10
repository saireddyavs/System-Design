package main

import (
	"chat-application/internal/broker"
	"chat-application/internal/repositories"
	"chat-application/internal/services"
	"fmt"
	"log"
	"time"
)

func main() {
	// Initialize repositories
	userRepo := repositories.NewInMemoryUserRepository()
	msgRepo := repositories.NewInMemoryMessageRepository()
	roomRepo := repositories.NewInMemoryChatRoomRepository()

	// Initialize broker with direct delivery strategy
	msgBroker := broker.NewInMemoryBroker(&broker.DirectDelivery{})

	// Initialize services
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	chatService := services.NewChatService(roomRepo, userRepo)
	msgFactory := services.NewMessageFactory()
	msgService := services.NewMessageService(msgRepo, roomRepo, msgBroker, msgFactory)

	// Demo: Register users
	user1, err := authService.Register("alice", "alice@example.com", "password123")
	if err != nil {
		log.Fatal("Register alice:", err)
	}
	user2, err := authService.Register("bob", "bob@example.com", "password123")
	if err != nil {
		log.Fatal("Register bob:", err)
	}

	// Demo: Login
	loggedIn, err := authService.Login("alice", "password123")
	if err != nil {
		log.Fatal("Login:", err)
	}
	fmt.Printf("Logged in: %s\n", loggedIn.Username)

	// Demo: Set online
	_ = userService.SetOnline(user1.ID)

	// Demo: Create 1:1 room
	room, err := chatService.CreateOneOnOneRoom(user1.ID, user2.ID)
	if err != nil {
		log.Fatal("Create room:", err)
	}
	fmt.Printf("Created room: %s\n", room.ID)

	// Demo: Subscribe bob for real-time messages
	msgChan := msgService.Subscribe(user2.ID)
	go func() {
		for msg := range msgChan {
			fmt.Printf("[Real-time] Bob received: %s from %s\n", msg.Content, msg.SenderID)
		}
	}()

	// Demo: Send message
	msg, err := msgService.SendMessage(user1.ID, room.ID, "Hello Bob!", "")
	if err != nil {
		log.Fatal("Send message:", err)
	}
	fmt.Printf("Sent message: %s\n", msg.Content)

	// Demo: Create group
	groupRoom, err := chatService.CreateGroupRoom(user1.ID, "Team Chat", []string{user2.ID})
	if err != nil {
		log.Fatal("Create group:", err)
	}
	fmt.Printf("Created group: %s\n", groupRoom.Name)

	// Demo: Send to group
	_, _ = msgService.SendMessage(user1.ID, groupRoom.ID, "Welcome to the team!", "")

	// Give time for async delivery
	time.Sleep(100 * time.Millisecond)

	// Demo: Get message history
	history, err := msgService.GetMessageHistory(room.ID, user1.ID, 10, 0)
	if err != nil {
		log.Fatal("Get history:", err)
	}
	fmt.Printf("Message history: %d messages\n", len(history))

	// Cleanup
	msgService.Unsubscribe(user2.ID)

	fmt.Println("\nChat application demo completed successfully!")
}
