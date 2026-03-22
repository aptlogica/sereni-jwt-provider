// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"strings"
	"testing"
	"time"

	sereni "github.com/aptlogica/sereni-jwt-provider"
)

func TestNewProvider(t *testing.T) {
	config := sereni.Config{
		Secret:    "test-secret",
		Expiry:    time.Hour,
		Issuer:    "test-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

func TestProvider_GenerateToken(t *testing.T) {
	tests := []struct {
		name        string
		config      sereni.Config
		claims      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "success - HS256 algorithm",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			claims: map[string]interface{}{
				"user_id": "123",
				"email":   "test@example.com",
			},
			expectError: false,
		},
		{
			name: "success - HS384 algorithm",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS384",
			},
			claims: map[string]interface{}{
				"user_id": "456",
			},
			expectError: false,
		},
		{
			name: "success - HS512 algorithm",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS512",
			},
			claims: map[string]interface{}{
				"user_id": "789",
			},
			expectError: false,
		},
		{
			name: "success - default algorithm when empty",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "",
			},
			claims: map[string]interface{}{
				"user_id": "abc",
			},
			expectError: false,
		},
		{
			name: "failure - empty secret",
			config: sereni.Config{
				Secret:    "",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			claims: map[string]interface{}{
				"user_id": "123",
			},
			expectError: true,
			errorMsg:    "secret key is required",
		},
		{
			name: "success - empty claims",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			claims:      map[string]interface{}{},
			expectError: false,
		},
		{
			name: "success - nil claims",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			claims:      nil,
			expectError: false,
		},
		{
			name: "success - complex claims",
			config: sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			claims: map[string]interface{}{
				"user_id":        "123",
				"email":          "test@example.com",
				"roles":          []string{"admin", "user"},
				"custom_int":     42,
				"custom_bool":    true,
				"custom_float":   3.14,
				"email_verified": true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := sereni.NewProvider(tt.config)
			token, err := provider.GenerateToken(tt.claims)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != nil && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if token != "" {
					t.Errorf("expected empty token on error, got %s", token)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token == "" {
					t.Error("expected non-empty token")
				}
				// JWT tokens have 3 parts separated by dots
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("expected JWT with 3 parts, got %d parts", len(parts))
				}
			}
		})
	}
}

func TestProvider_ValidateToken(t *testing.T) {
	config := sereni.Config{
		Secret:    "test-secret-key",
		Expiry:    time.Hour,
		Issuer:    "test-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)

	// Generate a valid token first
	claims := map[string]interface{}{
		"user_id": "123",
		"email":   "test@example.com",
		"roles":   "admin,user",
	}
	validToken, err := provider.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		config      sereni.Config
		token       string
		expectError bool
		checkClaims bool
	}{
		{
			name:        "success - valid token",
			config:      config,
			token:       validToken,
			expectError: false,
			checkClaims: true,
		},
		{
			name:        "failure - invalid token",
			config:      config,
			token:       "invalid.token.here",
			expectError: true,
		},
		{
			name:        "failure - empty token",
			config:      config,
			token:       "",
			expectError: true,
		},
		{
			name:        "failure - malformed token",
			config:      config,
			token:       "not-a-jwt",
			expectError: true,
		},
		{
			name: "failure - empty secret",
			config: sereni.Config{
				Secret:    "",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			token:       validToken,
			expectError: true,
		},
		{
			name: "failure - wrong secret",
			config: sereni.Config{
				Secret:    "wrong-secret",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: "HS256",
			},
			token:       validToken,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := sereni.NewProvider(tt.config)
			result, err := p.ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if result != nil {
					t.Error("expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if tt.checkClaims {
					if result["user_id"] != "123" {
						t.Errorf("expected user_id '123', got '%v'", result["user_id"])
					}
					if result["email"] != "test@example.com" {
						t.Errorf("expected email 'test@example.com', got '%v'", result["email"])
					}
				}
			}
		})
	}
}

func TestProvider_ValidateToken_DifferentAlgorithms(t *testing.T) {
	algorithms := []string{"HS256", "HS384", "HS512"}

	for _, algo := range algorithms {
		t.Run("algorithm_"+algo, func(t *testing.T) {
			config := sereni.Config{
				Secret:    "test-secret-key",
				Expiry:    time.Hour,
				Issuer:    "test-issuer",
				Algorithm: algo,
			}

			provider := sereni.NewProvider(config)
			claims := map[string]interface{}{
				"user_id": "123",
			}

			token, err := provider.GenerateToken(claims)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			result, err := provider.ValidateToken(token)
			if err != nil {
				t.Errorf("failed to validate token: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestProvider_RefreshToken(t *testing.T) {
	config := sereni.Config{
		Secret:    "test-secret-key",
		Expiry:    time.Hour,
		Issuer:    "test-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)

	claims := map[string]interface{}{
		"user_id": "123",
		"email":   "test@example.com",
		"roles":   "admin",
	}

	originalToken, err := provider.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "success - valid token",
			token:       originalToken,
			expectError: false,
		},
		{
			name:        "failure - invalid token",
			token:       "invalid.token.here",
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
			newToken, err := provider.RefreshToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if newToken != "" {
					t.Error("expected empty token on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if newToken == "" {
					t.Error("expected non-empty token")
				}
				// Note: We don't check if tokens are different because if generated
				// in the same second, they will be identical (same iat, exp, claims)

				// Verify new token contains the same claims
				newClaims, err := provider.ValidateToken(newToken)
				if err != nil {
					t.Errorf("failed to validate new token: %v", err)
				}
				if newClaims["user_id"] != "123" {
					t.Errorf("expected user_id '123', got '%v'", newClaims["user_id"])
				}
				if newClaims["email"] != "test@example.com" {
					t.Errorf("expected email 'test@example.com', got '%v'", newClaims["email"])
				}
			}
		})
	}
}

func TestProvider_TokenExpiry(t *testing.T) {
	// Create a provider with 1 second expiry
	config := sereni.Config{
		Secret:    "test-secret-key",
		Expiry:    time.Second,
		Issuer:    "test-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)
	claims := map[string]interface{}{
		"user_id": "123",
	}

	token, err := provider.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = provider.ValidateToken(token)
	if err != nil {
		t.Errorf("token should be valid immediately: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should now be invalid
	_, err = provider.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestProvider_IssuerClaim(t *testing.T) {
	config := sereni.Config{
		Secret:    "test-secret-key",
		Expiry:    time.Hour,
		Issuer:    "custom-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)
	claims := map[string]interface{}{
		"user_id": "123",
	}

	token, err := provider.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	result, err := provider.ValidateToken(token)
	if err != nil {
		t.Errorf("failed to validate token: %v", err)
	}

	if result["iss"] != "custom-issuer" {
		t.Errorf("expected issuer 'custom-issuer', got '%v'", result["iss"])
	}
}

func TestProvider_StandardClaims(t *testing.T) {
	config := sereni.Config{
		Secret:    "test-secret-key",
		Expiry:    time.Hour,
		Issuer:    "test-issuer",
		Algorithm: "HS256",
	}

	provider := sereni.NewProvider(config)
	claims := map[string]interface{}{
		"user_id": "123",
	}

	beforeGenerate := time.Now().Unix()
	token, err := provider.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	afterGenerate := time.Now().Unix()

	result, err := provider.ValidateToken(token)
	if err != nil {
		t.Errorf("failed to validate token: %v", err)
	}

	// Check iat claim
	iat, ok := result["iat"].(float64)
	if !ok {
		t.Fatal("iat claim not found or not a number")
	}
	if int64(iat) < beforeGenerate || int64(iat) > afterGenerate {
		t.Error("iat claim is not within expected range")
	}

	// Check exp claim
	exp, ok := result["exp"].(float64)
	if !ok {
		t.Fatal("exp claim not found or not a number")
	}
	expectedExp := int64(iat) + 3600 // 1 hour
	if int64(exp) != expectedExp {
		t.Errorf("expected exp %d, got %d", expectedExp, int64(exp))
	}
}
