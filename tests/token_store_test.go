package tests

import (
	"auth-service/internal/repository"
	"fmt"
	"testing"
	"time"
)

func TestNewTokenStore(t *testing.T) {
	store := repository.NewTokenStore()

	if store == nil {
		t.Error("expected store to be created")
	}
}

func TestTokenStore_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		email       string
		password    string
		roles       []string
		expectError bool
	}{
		{
			name:        "success - valid user creation with custom ID",
			userID:      "custom-user-123",
			email:       "test@example.com",
			password:    "hashedpass",
			roles:       []string{"user"},
			expectError: false,
		},
		{
			name:        "success - empty userID generates UUID",
			userID:      "",
			email:       "test2@example.com",
			password:    "hashedpass",
			roles:       nil,
			expectError: false,
		},
		{
			name:        "failure - duplicate email",
			userID:      "another-user-456",
			email:       "test@example.com",
			password:    "hashedpass2",
			roles:       []string{"admin"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewTokenStore()

			// Create first user for duplicate test
			if tt.expectError && tt.name == "failure - duplicate email" {
				store.CreateUser("first-user-id", "test@example.com", "hashedpass", []string{"user"})
			}

			user, err := store.CreateUser(tt.userID, tt.email, tt.password, tt.roles)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if user != nil {
					t.Error("expected nil user on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if user == nil {
					t.Fatalf("expected user, got nil")
				}
				if user.Email != tt.email {
					t.Errorf("expected email %s, got %s", tt.email, user.Email)
				}
				if user.ID == "" {
					t.Error("expected non-empty user ID")
				}
				// Check if custom userID was used
				if tt.userID != "" && user.ID != tt.userID {
					t.Errorf("expected user ID %s, got %s", tt.userID, user.ID)
				}
				expectedRoles := tt.roles
				if tt.roles == nil || len(tt.roles) == 0 {
					expectedRoles = []string{"user"}
				}
				if len(user.Roles) != len(expectedRoles) || user.Roles[0] != expectedRoles[0] {
					t.Errorf("expected roles %v, got %v", expectedRoles, user.Roles)
				}
			}
		})
	}
}

func TestTokenStore_GetUserByEmail(t *testing.T) {
	store := repository.NewTokenStore()

	// Create user
	user, err := store.CreateUser("test-user-id", "test@example.com", "hashedpass", []string{"user"})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Get user - success case
	retrieved, err := store.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, retrieved.ID)
	}

	// Get user - not found case
	_, err = store.GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestTokenStore_StoreRefreshToken(t *testing.T) {
	store := repository.NewTokenStore()

	// Test basic storage
	err := store.StoreRefreshToken("user-id", "token1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, err := store.ValidateRefreshToken("token1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user-id" {
		t.Errorf("expected user-id, got %s", userID)
	}

	// Test token cleanup (store 6 tokens, should keep only 5)
	for i := 2; i <= 6; i++ {
		token := fmt.Sprintf("token%d", i)
		err := store.StoreRefreshToken("user-id", token, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("unexpected error storing %s: %v", token, err)
		}
	}

	// First token should be cleaned up
	_, err = store.ValidateRefreshToken("token1")
	if err == nil {
		t.Error("expected token1 to be cleaned up")
	}

	// Last 5 tokens should still be valid
	for i := 2; i <= 6; i++ {
		token := fmt.Sprintf("token%d", i)
		userID, err := store.ValidateRefreshToken(token)
		if err != nil {
			t.Fatalf("token %s should still be valid: %v", token, err)
		}
		if userID != "user-id" {
			t.Errorf("expected user-id for %s, got %s", token, userID)
		}
	}
}

func TestTokenStore_RevokeRefreshToken(t *testing.T) {
	store := repository.NewTokenStore()

	err := store.StoreRefreshToken("user-id", "token", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to store token: %v", err)
	}

	err = store.RevokeRefreshToken("token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.ValidateRefreshToken("token")
	if err == nil {
		t.Error("expected error after revocation")
	}
}

func TestTokenStore_CleanExpiredTokens(t *testing.T) {
	store := repository.NewTokenStore()

	// Store expired token
	err := store.StoreRefreshToken("user-id", "expired", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("failed to store expired token: %v", err)
	}

	store.CleanExpiredTokens()

	_, err = store.ValidateRefreshToken("expired")
	if err == nil {
		t.Error("expected expired token to be cleaned")
	}
}

func TestTokenStore_GetUserByID(t *testing.T) {
	store := repository.NewTokenStore()

	// Create user
	user, err := store.CreateUser("test-user-id", "test@example.com", "hashedpass", []string{"user"})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Get user by ID - success case
	retrieved, err := store.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, retrieved.ID)
	}

	// Get user by ID - not found case
	_, err = store.GetUserByID("nonexistent-id")
	if err == nil {
		t.Error("expected error for non-existent user ID")
	}
}

func TestTokenStore_RevokeAllUserTokens(t *testing.T) {
	store := repository.NewTokenStore()

	userID := "user-id"

	// Store multiple tokens
	err := store.StoreRefreshToken(userID, "token1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to store token1: %v", err)
	}
	err = store.StoreRefreshToken(userID, "token2", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to store token2: %v", err)
	}

	// Revoke all
	err = store.RevokeAllUserTokens(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check tokens are revoked
	_, err = store.ValidateRefreshToken("token1")
	if err == nil {
		t.Error("expected token1 to be revoked")
	}
	_, err = store.ValidateRefreshToken("token2")
	if err == nil {
		t.Error("expected token2 to be revoked")
	}
}

func TestTokenStore_GetStats(t *testing.T) {
	store := repository.NewTokenStore()

	// Create users and tokens
	_, err := store.CreateUser("user1-id", "user1@example.com", "pass", []string{"user"})
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	_, err = store.CreateUser("user2-id", "user2@example.com", "pass", []string{"user"})
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	err = store.StoreRefreshToken("user1-id", "token1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to store token1: %v", err)
	}

	stats := store.GetStats()

	if stats["users"] != 2 {
		t.Errorf("expected 2 users, got %d", stats["users"])
	}
	if stats["refresh_tokens"] != 1 {
		t.Errorf("expected 1 refresh token, got %d", stats["refresh_tokens"])
	}
}

func TestTokenStore_StartCleanupRoutine(t *testing.T) {
	store := repository.NewTokenStore()

	// First call should start the routine
	store.StartCleanupRoutine()

	// Second call should not start another routine
	store.StartCleanupRoutine()

	store.Close()
}

func TestTokenStore_Close(t *testing.T) {
	store := repository.NewTokenStore()

	// Close without starting should do nothing
	store.Close()
}
