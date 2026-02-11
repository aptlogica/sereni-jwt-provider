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
		userID      string
		email       string
		password    string
		roles       []string
		expectError bool
	}{
		{
			name:        "success - valid registration",
			userID:      "test-user-123",
			email:       "test@example.com",
			password:    "password123",
			roles:       []string{"user"},
			expectError: false,
		},
		{
			name:        "failure - duplicate email",
			userID:      "test-user-456",
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
				_, _ = service.Register("existing-user-id", "test@example.com", "password123", []string{"user"})
			}

			user, err := service.Register(tt.userID, tt.email, tt.password, tt.roles)
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
				if user.ID != tt.userID {
					t.Errorf("expected user ID %s, got %s", tt.userID, user.ID)
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
		setupUser   bool
	}{
		{
			name:        "success - valid login",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
			setupUser:   true,
		},
		{
			name:        "success - login with any password (password validation skipped)",
			email:       "test@example.com",
			password:    "anypassword",
			expectError: false,
			setupUser:   true,
		},
		{
			name:        "failure - user not found",
			email:       "nonexistent@example.com",
			password:    "password123",
			expectError: true,
			setupUser:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			// Register user for tests that need it
			if tt.setupUser {
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
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
				if tokens.TokenType != "Bearer" {
					t.Errorf("expected token type Bearer, got %s", tokens.TokenType)
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
		tokenType   string
	}{
		{
			name:        "success - valid access token",
			token:       "valid-token",
			expectError: false,
			tokenType:   "access",
		},
		{
			name:        "success - valid refresh token",
			token:       "valid-refresh-token",
			expectError: false,
			tokenType:   "refresh",
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
		{
			name:        "failure - malformed token",
			token:       "not.a.valid.jwt.token",
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
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user", "admin"})
				tokens, _ := service.Login("test@example.com", "password123")
				if tt.tokenType == "refresh" {
					validToken = tokens.RefreshToken
				} else {
					validToken = tokens.AccessToken
				}
			}

			tokenToValidate := tt.token
			if !tt.expectError {
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
				if claims.TokenType != tt.tokenType {
					t.Errorf("expected token type %s, got %s", tt.tokenType, claims.TokenType)
				}
				if len(claims.Roles) == 0 {
					t.Error("expected roles in claims")
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
				user, _ := service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
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
	}{
		{
			name: "success - valid refresh token",
			setupFunc: func(service *services.JWTService) string {
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
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
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				return tokens.AccessToken // Using access token instead of refresh
			},
			expectError: true,
		},
		{
			name: "failure - revoked refresh token",
			setupFunc: func(service *services.JWTService) string {
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				refreshToken := tokens.RefreshToken
				// Revoke the token
				_ = service.Logout(refreshToken)
				return refreshToken
			},
			expectError: true,
		},
		{
			name:        "failure - empty refresh token",
			setupFunc:   func(service *services.JWTService) string { return "" },
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
					t.Errorf("expected error, got nil. Token: %s", tokenToRefresh)
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
	tests := []struct {
		name        string
		setupFunc   func(*services.JWTService) string
		expectError bool
	}{
		{
			name: "success - logout with valid refresh token",
			setupFunc: func(service *services.JWTService) string {
				_, _ = service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
				tokens, _ := service.Login("test@example.com", "password123")
				return tokens.RefreshToken
			},
			expectError: false,
		},
		{
			name:        "failure - logout with invalid token (error expected)",
			setupFunc:   func(service *services.JWTService) string { return "invalid-token" },
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			refreshToken := tt.setupFunc(service)

			// Logout
			err := service.Logout(refreshToken)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// If we had a valid token, verify it's revoked
				if tt.name == "success - logout with valid refresh token" {
					_, err := service.RefreshAccessToken(refreshToken)
					if err == nil {
						t.Error("expected error when refreshing revoked token")
					}
				}
			}
		})
	}
}

func TestJWTService_LogoutAll(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "success - logout all user tokens",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			service := services.NewJWTService("test-secret", tokenStore)

			// Register and create multiple login sessions
			user, err := service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
			if err != nil {
				t.Fatalf("failed to register: %v", err)
			}

			tokens1, _ := service.Login("test@example.com", "password123")
			tokens2, _ := service.Login("test@example.com", "password123")

			// Logout all
			err = service.LogoutAll(user.ID)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify both tokens are revoked
				_, err1 := service.RefreshAccessToken(tokens1.RefreshToken)
				_, err2 := service.RefreshAccessToken(tokens2.RefreshToken)
				if err1 == nil || err2 == nil {
					t.Error("expected both tokens to be revoked")
				}
			}
		})
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	tokenStore := repository.NewTokenStore()
	service := services.NewJWTService("test-secret", tokenStore)

	// Create a user
	user, err := service.Register("test-user-id", "test@example.com", "password123", []string{"user"})
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
