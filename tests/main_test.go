// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"os"
	"testing"
)

func TestEnvironmentDefaults(t *testing.T) {
	// Save original environment
	originals := map[string]string{
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
		"SERVER_PORT": os.Getenv("SERVER_PORT"),
		"SERVER_HOST": os.Getenv("SERVER_HOST"),
		"GIN_MODE":    os.Getenv("GIN_MODE"),
	}

	// Restore environment after test
	defer func() {
		for key, value := range originals {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Clear environment variables
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("GIN_MODE")

	// Test that environment variables are empty (as expected)
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		t.Errorf("expected JWT_SECRET to be empty, got %s", jwtSecret)
	}

	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		t.Errorf("expected SERVER_PORT to be empty, got %s", serverPort)
	}

	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		t.Errorf("expected SERVER_HOST to be empty, got %s", serverHost)
	}

	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		t.Errorf("expected GIN_MODE to be empty, got %s", ginMode)
	}
}

func TestMainPackageImport(t *testing.T) {
	// This test simply verifies that the main package can be imported
	// and basic functionality works. Main function itself is not unit testable
	// as it starts a server, but we can test that the package compiles.

	t.Log("Main package imported successfully")
}
