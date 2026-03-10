package tests

import (
	"chat-application/internal/repositories"
	"chat-application/internal/services"
	"testing"
)

func TestAuthService_Register(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	user, err := authService.Register("alice", "alice@test.com", "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("expected email alice@test.com, got %s", user.Email)
	}
	if user.ID == "" {
		t.Error("expected non-empty user ID")
	}
	if user.PasswordHash == "password123" {
		t.Error("password should be hashed")
	}
}

func TestAuthService_RegisterDuplicateUsername(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	_, _ = authService.Register("alice", "alice@test.com", "password123")
	_, err := authService.Register("alice", "bob@test.com", "password456")
	if err != services.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	_, _ = authService.Register("alice", "alice@test.com", "password123")
	_, err := authService.Register("bob", "alice@test.com", "password456")
	if err != services.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	_, _ = authService.Register("alice", "alice@test.com", "password123")

	user, err := authService.Login("alice", "password123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}
}

func TestAuthService_LoginInvalidPassword(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	_, _ = authService.Register("alice", "alice@test.com", "password123")

	_, err := authService.Login("alice", "wrongpassword")
	if err != services.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_LoginInvalidUsername(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	authService := services.NewAuthService(userRepo)

	_, err := authService.Login("nonexistent", "password123")
	if err != services.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
