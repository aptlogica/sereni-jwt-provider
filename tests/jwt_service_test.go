// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"github.com/aptlogica/sereni-jwt-provider/internal/services"
	"testing"
	"time"
)

func TestNewJWTService(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)
	if service == nil {
		t.Fatal("expected service to be created")
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := services.NewJWTService("test-secret", 120, 3600)
	user := &models.User{
		ID:             "test-user-id",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user", "admin"},
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
	if tokens.ExpiresIn != 120 {
		t.Errorf("expected expires in 120, got %d", tokens.ExpiresIn)
	}
}

func TestJWTService_Login(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)

	req := &models.LoginRequest{
		ID:             "test-user-id",
		Email:          "test@example.com",
		Roles:          []string{"user"},
		EMAIL_VERIFIED: true,
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
	service := services.NewJWTService("test-secret", 900, 604800)
	user := &models.User{
		ID:             "test-user-id",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user", "admin"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token, true)
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
			if !claims.EMAIL_VERIFIED {
				t.Error("expected email_verified=true in claims")
			}
		})
	}
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)
	user := &models.User{
		ID:             "test-user-id",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
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
			newTokens, err := service.RefreshAccessToken(tt.refreshToken)
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

func TestJWTService_ExpiredToken(t *testing.T) {
	service := services.NewJWTService("test-secret", -1, 3600)
	user := &models.User{
		ID:    "expired-user",
		Email: "expired@example.com",
		Roles: []string{"user"},
	}

	pair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("failed generating token pair: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	claims, err := service.ValidateToken(pair.AccessToken, true)
	if err == nil {
		t.Fatal("expected expired token error, got nil")
	}
	if claims != nil {
		t.Fatal("expected nil claims for expired token")
	}
}

func TestJWTService_GenerateTokenPair_EmptyRoles(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)
	user := &models.User{
		ID:             "no-roles-user",
		Email:          "noroles@example.com",
		EMAIL_VERIFIED: false,
		Roles:          []string{},
	}

	tokens, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens, got nil")
	}

	// Validate the token
	claims, err := service.ValidateToken(tokens.AccessToken, true)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.Roles != "" {
		t.Errorf("expected empty roles string, got %s", claims.Roles)
	}
	if claims.EMAIL_VERIFIED != false {
		t.Error("expected email_verified to be false")
	}
}

func TestJWTService_GenerateTokenPair_SingleRole(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)
	user := &models.User{
		ID:             "single-role-user",
		Email:          "single@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"admin"},
	}

	tokens, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := service.ValidateToken(tokens.AccessToken, true)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.Roles != "admin" {
		t.Errorf("expected roles 'admin', got %s", claims.Roles)
	}
}

func TestJWTService_ValidateToken_MalformedToken(t *testing.T) {
	service := services.NewJWTService("test-secret", 900, 604800)

	tokens, err := service.GenerateTokenPair(&models.User{
		ID:             "test-user-id",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	})
	if err != nil {
		t.Fatalf("failed generating token pair: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed token format",
			token: "this.is.not.a.valid.jwt.token",
		},
		{
			name:  "incomplete token",
			token: "incomplete",
		},
		{
			name:  "token with wrong signature",
			token: tokens.AccessToken + "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token, true)
			if err == nil {
				t.Error("expected error for malformed token")
			}
			if claims != nil {
				t.Error("expected nil claims for malformed token")
			}
		})
	}
}

func TestJWTService_RefreshAccessToken_ErrorCases(t *testing.T) {
	t.Run("expired refresh token", func(t *testing.T) {
		expiredService := services.NewJWTService("test-secret", 900, -1)
		user := &models.User{
			ID:             "expired-refresh-user",
			Email:          "expiredrefresh@example.com",
			EMAIL_VERIFIED: true,
			Roles:          []string{"user"},
		}

		tokens, err := expiredService.GenerateTokenPair(user)
		if err != nil {
			t.Fatalf("failed to generate tokens: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		newTokens, err := expiredService.RefreshAccessToken(tokens.RefreshToken)
		if err == nil {
			t.Error("expected error for expired refresh token")
		}
		if newTokens != nil {
			t.Error("expected nil tokens for expired refresh token")
		}
	})
}
