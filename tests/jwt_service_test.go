package tests

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"fmt"
	"testing"
	"time"
)

// mockTokenStore simulates StoreRefreshToken failure
type mockTokenStore struct {
	repository.TokenStore
	failStore bool
}

func (m *mockTokenStore) StoreRefreshToken(userID, token string, expiry time.Time) error {
	if m.failStore {
		return fmt.Errorf("store error")
	}
	return nil
}

func TestJWTService_GenerateTokenPair_ErrorPaths(t *testing.T) {
	service := services.NewJWTService("test-secret", &mockTokenStore{failStore: true})
	user := &models.User{ID: "id", Email: "e", Roles: []string{"user"}}

	_, err := service.GenerateTokenPair(user)
	if err == nil {
		t.Error("expected error from StoreRefreshToken failure")
	}
}

func TestNewJWTService(t *testing.T) {
	secretKey := "test-secret"
	tokenStore := repository.NewTokenStore()

	service := services.NewJWTService(secretKey, tokenStore)
	if service == nil {
		t.Errorf("expected service to be created")
	}
}

func TestJWTService_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		roles       []string
		expectError bool
	}{
		{
			name:        "success - valid registration",
			email:       "test@example.com",
			password:    "password123",
			roles:       []string{"user"},
			expectError: false,
		},
		{
			name:        "failure - duplicate email",
			email:       "test@example.com",
			password:    "password456",
			roles:       []string{"admin"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			// Register first user for duplicate test
			if tt.expectError {
				_, _ = service.Register("test@example.com", "password123", []string{"user"})
			}

			user, err := service.Register(tt.email, tt.password, tt.roles)
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
			}
		})
	}
}

func TestJWTService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "success - valid login",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "failure - wrong password",
			email:       "test@example.com",
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "failure - user not found",
			email:       "nonexistent@example.com",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			// Register user for valid login and wrong password tests
			if tt.name != "failure - user not found" {
				_, _ = service.Register("test@example.com", "password123", []string{"user"})
			}

			tokens, err := service.Login(tt.email, tt.password)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tokens != nil {
					t.Error("expected nil tokens on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tokens == nil {
					t.Fatalf("expected tokens, got nil")
				}
				if tokens.AccessToken == "" {
					t.Errorf("expected non-empty access token")
				}
				if tokens.RefreshToken == "" {
					t.Errorf("expected non-empty refresh token")
				}
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "success - valid token",
			token:       "valid-token",
			expectError: false,
		},
		{
			name:        "failure - invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "failure - empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			var validToken string
			if !tt.expectError {
				// Setup valid token
				_, _ = service.Register("test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				validToken = tokens.AccessToken
			}

			tokenToValidate := tt.token
			if tt.name == "success - valid token" {
				tokenToValidate = validToken
			}

			claims, err := service.ValidateToken(tokenToValidate)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if claims != nil {
					t.Error("expected nil claims on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if claims == nil {
					t.Fatalf("expected claims, got nil")
				}
				if claims.Email != "test@example.com" {
					t.Errorf("expected email test@example.com, got %s", claims.Email)
				}
			}
		})
	}
}

func TestJWTService_GetUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		expectError bool
	}{
		{
			name:        "success - valid user",
			userID:      "valid-user-id",
			expectError: false,
		},
		{
			name:        "failure - user not found",
			userID:      "nonexistent-user",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			var validUserID string
			if !tt.expectError {
				// Setup user
				user, _ := service.Register("test@example.com", "password123", []string{"user"})
				validUserID = user.ID
			}

			userID := tt.userID
			if tt.name == "success - valid user" {
				userID = validUserID
			}

			profile, err := service.GetUserProfile(userID)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if profile != nil {
					t.Error("expected nil profile on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if profile == nil {
					t.Fatalf("expected profile, got nil")
				}
				if profile.Email != "test@example.com" {
					t.Errorf("expected email test@example.com, got %s", profile.Email)
				}
			}
		})
	}
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*services.JWTService) string
		expectError bool
		errorType   error
	}{
		{
			name: "success - valid refresh token",
			setupFunc: func(service *services.JWTService) string {
				_, _ = service.Register("test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				return tokens.RefreshToken
			},
			expectError: false,
		},
		{
			name:        "failure - invalid refresh token",
			setupFunc:   func(service *services.JWTService) string { return "invalid-token" },
			expectError: true,
		},
		{
			name: "failure - access token used as refresh token",
			setupFunc: func(service *services.JWTService) string {
				_, _ = service.Register("test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				return tokens.AccessToken // Using access token instead of refresh
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			tokenToRefresh := tt.setupFunc(service)

			newTokens, err := service.RefreshAccessToken(tokenToRefresh)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if newTokens != nil {
					t.Error("expected nil tokens on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if newTokens == nil {
					t.Fatalf("expected new tokens, got nil")
				}
				if newTokens.AccessToken == "" {
					t.Errorf("expected non-empty access token")
				}
				if newTokens.RefreshToken == "" {
					t.Errorf("expected non-empty refresh token")
				}
			}
		})
	}
}

func TestJWTService_Logout(t *testing.T) {
	tokenStore := repository.NewTokenStore()
	service := services.NewJWTService("test-secret", tokenStore)

	// Register and login
	_, err := service.Register("test@example.com", "password123", []string{"user"})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	tokens, err := service.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// Logout
	err = service.Logout(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to refresh with revoked token
	_, err = service.RefreshAccessToken(tokens.RefreshToken)
	if err == nil {
		t.Error("expected error when refreshing revoked token")
	}
}

func TestJWTService_LogoutAll(t *testing.T) {
	tokenStore := repository.NewTokenStore()
	service := services.NewJWTService("test-secret", tokenStore)

	// Register and login
	user, err := service.Register("test@example.com", "password123", []string{"user"})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	_, err = service.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// Logout all
	err = service.LogoutAll(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	tokenStore := repository.NewTokenStore()
	service := services.NewJWTService("test-secret", tokenStore)

	// Create a user
	user, err := service.Register("test@example.com", "password123", []string{"user"})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// Generate token pair directly
	tokens, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil {
		t.Fatalf("expected tokens, got nil")
	}
	if tokens.AccessToken == "" {
		t.Errorf("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Errorf("expected non-empty refresh token")
	}
	if tokens.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", tokens.TokenType)
	}
	if tokens.ExpiresIn != 900 { // 15 minutes in seconds
		t.Errorf("expected expires in 900, got %d", tokens.ExpiresIn)
	}
}
