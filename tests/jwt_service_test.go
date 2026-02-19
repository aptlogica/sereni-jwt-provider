package tests

import (
	"auth-service/internal/models"
	"auth-service/internal/services"
	"testing"
)

func TestNewJWTService(t *testing.T) {
	service := services.NewJWTService("test-secret")
	if service == nil {
		t.Fatal("expected service to be created")
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := services.NewJWTService("test-secret")
	user := &models.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Password: "password123",
		Roles:    []string{"user", "admin"},
	}

	tokens, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens, got nil")
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if tokens.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", tokens.TokenType)
	}
	if tokens.ExpiresIn != int64(services.AccessTokenDuration.Seconds()) {
		t.Errorf("expected expires in %d, got %d", int64(services.AccessTokenDuration.Seconds()), tokens.ExpiresIn)
	}
}

func TestJWTService_Login(t *testing.T) {
	service := services.NewJWTService("test-secret")

	req := &models.LoginRequest{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Password: "password123",
		Roles:    []string{"user"},
	}

	tokens, err := service.Login(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens, got nil")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected both access and refresh tokens")
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	service := services.NewJWTService("test-secret")
	user := &models.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Password: "password123",
		Roles:    []string{"user", "admin"},
	}

	pair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("failed generating token pair: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		tokenType   string
	}{
		{
			name:        "success - valid access token",
			token:       pair.AccessToken,
			expectError: false,
			tokenType:   services.TokenTypeAccess,
		},
		{
			name:        "success - valid refresh token",
			token:       pair.RefreshToken,
			expectError: false,
			tokenType:   services.TokenTypeRefresh,
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
			claims, err := service.ValidateToken(tt.token)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if claims != nil {
					t.Error("expected nil claims on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims == nil {
				t.Fatal("expected claims, got nil")
			}
			if claims.UserID != user.ID {
				t.Errorf("expected user_id %s, got %s", user.ID, claims.UserID)
			}
			if claims.Email != user.Email {
				t.Errorf("expected email %s, got %s", user.Email, claims.Email)
			}
			if claims.TokenType != tt.tokenType {
				t.Errorf("expected token type %s, got %s", tt.tokenType, claims.TokenType)
			}
			if claims.Roles != "user,admin" {
				t.Errorf("expected roles user,admin, got %s", claims.Roles)
			}
		})
	}
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	service := services.NewJWTService("test-secret")
	user := &models.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Password: "password123",
		Roles:    []string{"user"},
	}

	pair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("failed generating token pair: %v", err)
	}

	tests := []struct {
		name         string
		refreshToken string
		expectError  bool
	}{
		{
			name:         "success - valid refresh token",
			refreshToken: pair.RefreshToken,
			expectError:  false,
		},
		{
			name:         "failure - invalid refresh token",
			refreshToken: "invalid-token",
			expectError:  true,
		},
		{
			name:         "failure - access token used as refresh token",
			refreshToken: pair.AccessToken,
			expectError:  true,
		},
		{
			name:         "failure - empty refresh token",
			refreshToken: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTokens, err := service.RefreshAccessToken(tt.refreshToken, user)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if newTokens != nil {
					t.Error("expected nil tokens on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if newTokens == nil {
				t.Fatal("expected new tokens, got nil")
			}
			if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
				t.Error("expected both access and refresh tokens")
			}
		})
	}
}
