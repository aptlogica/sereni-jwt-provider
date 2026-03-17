// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
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
	os.Setenv("JWT_SECRET", "test-jwt-secret-key")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ACCESS_TOKEN_DURATION", "1800")
	os.Setenv("REFRESH_TOKEN_DURATION", "86400")

	cfg := config.LoadConfig()

	if cfg.JWTSecret != "test-jwt-secret-key" {
		t.Errorf("expected JWTSecret test-jwt-secret-key, got %s", cfg.JWTSecret)
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
