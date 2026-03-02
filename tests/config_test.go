package tests

import (
	"auth-service/internal/config"
	"os"
	"testing"
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
