// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"os"
	"testing"

	"github.com/aptlogica/sereni-jwt-provider/internal/config"
)

func restoreEnv(t *testing.T, key, value string) {
	t.Helper()
	if value == "" {
		os.Unsetenv(key)
		return
	}
	os.Setenv(key, value)
}

func TestLoadConfig_DefaultDurations(t *testing.T) {
	oldAccess := os.Getenv("ACCESS_TOKEN_DURATION")
	oldRefresh := os.Getenv("REFRESH_TOKEN_DURATION")
	oldHost := os.Getenv("SERVER_HOST")
	oldPort := os.Getenv("SERVER_PORT")
	defer restoreEnv(t, "ACCESS_TOKEN_DURATION", oldAccess)
	defer restoreEnv(t, "REFRESH_TOKEN_DURATION", oldRefresh)
	defer restoreEnv(t, "SERVER_HOST", oldHost)
	defer restoreEnv(t, "SERVER_PORT", oldPort)

	os.Unsetenv("ACCESS_TOKEN_DURATION")
	os.Unsetenv("REFRESH_TOKEN_DURATION")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")

	cfg := config.LoadConfig()
	if cfg.AccessTokenDuration != 900 {
		t.Fatalf("expected default access duration 900, got %d", cfg.AccessTokenDuration)
	}
	if cfg.RefreshTokenDuration != 604800 {
		t.Fatalf("expected default refresh duration 604800, got %d", cfg.RefreshTokenDuration)
	}
	if cfg.ServerHost != "0.0.0.0" {
		t.Fatalf("expected default host 0.0.0.0, got %s", cfg.ServerHost)
	}
	if cfg.ServerPort != "8081" {
		t.Fatalf("expected default port 8081, got %s", cfg.ServerPort)
	}
}

func TestLoadConfig_ParsesDurations(t *testing.T) {
	oldAccess := os.Getenv("ACCESS_TOKEN_DURATION")
	oldRefresh := os.Getenv("REFRESH_TOKEN_DURATION")
	defer restoreEnv(t, "ACCESS_TOKEN_DURATION", oldAccess)
	defer restoreEnv(t, "REFRESH_TOKEN_DURATION", oldRefresh)

	os.Setenv("ACCESS_TOKEN_DURATION", "120")
	os.Setenv("REFRESH_TOKEN_DURATION", "3600")

	cfg := config.LoadConfig()
	if cfg.AccessTokenDuration != 120 {
		t.Fatalf("expected access duration 120, got %d", cfg.AccessTokenDuration)
	}
	if cfg.RefreshTokenDuration != 3600 {
		t.Fatalf("expected refresh duration 3600, got %d", cfg.RefreshTokenDuration)
	}
}

func TestLoadConfig_InvalidDurations(t *testing.T) {
	oldAccess := os.Getenv("ACCESS_TOKEN_DURATION")
	oldRefresh := os.Getenv("REFRESH_TOKEN_DURATION")
	defer restoreEnv(t, "ACCESS_TOKEN_DURATION", oldAccess)
	defer restoreEnv(t, "REFRESH_TOKEN_DURATION", oldRefresh)

	// Set invalid durations - should fallback to defaults
	os.Setenv("ACCESS_TOKEN_DURATION", "not-a-number")
	os.Setenv("REFRESH_TOKEN_DURATION", "invalid")

	cfg := config.LoadConfig()
	// Should use defaults when parsing fails
	if cfg.AccessTokenDuration != 900 {
		t.Fatalf("expected default access duration 900 on invalid input, got %d", cfg.AccessTokenDuration)
	}
	if cfg.RefreshTokenDuration != 604800 {
		t.Fatalf("expected default refresh duration 604800 on invalid input, got %d", cfg.RefreshTokenDuration)
	}
}

func TestLoadConfig_AllEnvironmentVariables(t *testing.T) {
	oldVars := map[string]string{
		"JWT_SECRET":             os.Getenv("JWT_SECRET"),
		"SERVER_PORT":            os.Getenv("SERVER_PORT"),
		"SERVER_HOST":            os.Getenv("SERVER_HOST"),
		"ALLOWED_ORIGINS":        os.Getenv("ALLOWED_ORIGINS"),
		"ENV":                    os.Getenv("ENV"),
		"LOG_LEVEL":              os.Getenv("LOG_LEVEL"),
		"ACCESS_TOKEN_DURATION":  os.Getenv("ACCESS_TOKEN_DURATION"),
		"REFRESH_TOKEN_DURATION": os.Getenv("REFRESH_TOKEN_DURATION"),
	}
	defer func() {
		for k, v := range oldVars {
			restoreEnv(t, k, v)
		}
	}()

	// Set all environment variables
	jwtKey := testSigningKey(t)
	os.Setenv("JWT_SECRET", jwtKey)
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ACCESS_TOKEN_DURATION", "1800")
	os.Setenv("REFRESH_TOKEN_DURATION", "86400")

	cfg := config.LoadConfig()

	if cfg.JWTSecret != jwtKey {
		t.Errorf("expected JWTSecret %s, got %s", jwtKey, cfg.JWTSecret)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort 9090, got %s", cfg.ServerPort)
	}
	if cfg.ServerHost != "localhost" {
		t.Errorf("expected ServerHost localhost, got %s", cfg.ServerHost)
	}
	if cfg.AllowedOrigins != "http://localhost:3000,http://localhost:8080" {
		t.Errorf("expected AllowedOrigins with multiple origins, got %s", cfg.AllowedOrigins)
	}
	if cfg.Env != "production" {
		t.Errorf("expected Env production, got %s", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel info, got %s", cfg.LogLevel)
	}
	if cfg.AccessTokenDuration != 1800 {
		t.Errorf("expected AccessTokenDuration 1800, got %d", cfg.AccessTokenDuration)
	}
	if cfg.RefreshTokenDuration != 86400 {
		t.Errorf("expected RefreshTokenDuration 86400, got %d", cfg.RefreshTokenDuration)
	}
}

// TestValidateJWTSecret_EmptySecret tests that empty secrets are rejected
func TestValidateJWTSecret_EmptySecret(t *testing.T) {
	err := config.ValidateJWTSecret("")
	if err == nil {
		t.Error("expected error for empty JWT_SECRET, got nil")
	}
}

// TestValidateJWTSecret_TooShort tests that secrets shorter than 32 characters are rejected
func TestValidateJWTSecret_TooShort(t *testing.T) {
	testCases := []string{
		"a",
		"abc",
		"secret",
		"changeme",
		"weak123",
		"shortpass",
		"12345678",
		"verylongbutnotlongenoughyet", // 28 chars
	}

	for _, secret := range testCases {
		t.Run("secret="+secret[:len(secret)/2+1], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err == nil {
				t.Errorf("expected error for secret with %d chars, got nil", len(secret))
			}
		})
	}
}

// TestValidateJWTSecret_ValidLength tests that secrets 32+ characters are accepted (if not weak pattern)
func TestValidateJWTSecret_ValidLength(t *testing.T) {
	testCases := []string{
		testSigningKey(t), // 64-char hex string
		"this_is_a_valid_secret_with_32_chars_minimum_len", // 52 chars
		"AbCdEfGhIjKlMnOpQrStUvWxYz123456",                 // 32 chars mixed case + digits
		"super_secure_jwt_secret_with_special_!@#$%",       // 40+ chars with special
	}

	for _, secret := range testCases {
		t.Run("valid_"+secret[:16], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err != nil {
				t.Errorf("expected no error for valid secret, got: %v", err)
			}
		})
	}
}

// TestValidateJWTSecret_AllSameCharacter tests that secrets with all same characters are rejected
func TestValidateJWTSecret_AllSameCharacter(t *testing.T) {
	testCases := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 34 'a's
		"1111111111111111111111111111111111", // 34 '1's
		"________________________________",   // 32 underscores
	}

	for _, secret := range testCases {
		t.Run("all_same_char_"+string(secret[0]), func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err == nil {
				t.Errorf("expected error for weak pattern (all same character), got nil")
			}
		})
	}
}

// TestValidateJWTSecret_LowCharacterVariety tests that secrets with very low character variety are rejected
func TestValidateJWTSecret_LowCharacterVariety(t *testing.T) {
	testCases := []string{
		"abcdefghijklmnopqrstuvwxyzabcdefgh",  // only lowercase
		"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGH",  // only uppercase
		"12345678901234567890123456789012345", // only digits
	}

	for _, secret := range testCases {
		t.Run("low_variety_"+secret[:8], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err == nil {
				t.Errorf("expected error for low character variety, got nil")
			}
		})
	}
}

// TestValidateJWTSecret_MixedVariety tests that base64 and similar mixed-case secrets are accepted
func TestValidateJWTSecret_MixedVariety(t *testing.T) {
	testCases := []string{
		"abcdefghijklmnopqrstuvwxyz12345678", // lowercase + digits (like base64)
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ12345678", // uppercase + digits (like base64)
		"AbCdEfGhIjKlMnOpQrStUvWxYz123456",   // mixed case + digits
	}

	for _, secret := range testCases {
		t.Run("mixed_variety_"+secret[:8], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err != nil {
				t.Errorf("expected no error for base64-like secret with mixed variety, got: %v", err)
			}
		})
	}
}

// TestValidateJWTSecret_HighCharacterVariety tests that secrets with high character variety are accepted
func TestValidateJWTSecret_HighCharacterVariety(t *testing.T) {
	testCases := []string{
		"AbCdEfGhIjKlMnOpQrStUvWxYz123456",  // mixed case + digits
		"Aa1!_-+Bb2@#$Cc3%^&Dd4*(e5)fFgGhH", // mixed with special chars
		"MySecretKey!WithNumbers123AndCaps", // realistic looking secret
	}

	for _, secret := range testCases {
		t.Run("high_variety_"+secret[:8], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err != nil {
				t.Errorf("expected no error for high character variety, got: %v", err)
			}
		})
	}
}

// TestValidateJWTSecret_Base64Encoded tests that base64-encoded secrets are accepted
func TestValidateJWTSecret_Base64Encoded(t *testing.T) {
	// Base64 strings contain both upper and lower case letters + digits, meeting variety requirement
	testCases := []string{
		"dGhpc2lzYXZhbGlkYmFzZTY0c3RyaW5nd2l0aHBhZGRpbmdAQUEA", // valid base64
		"SGVsbG9Xb3JsZFdpdGhCYXNlNjRFbmNvZGluZ0FuZFBhZGRpbmc=", // with padding
	}

	for _, secret := range testCases {
		t.Run("base64_"+secret[:12], func(t *testing.T) {
			err := config.ValidateJWTSecret(secret)
			if err != nil {
				t.Errorf("expected no error for base64-encoded secret, got: %v", err)
			}
		})
	}
}
