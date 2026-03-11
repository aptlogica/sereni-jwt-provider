// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"auth-service/internal/utils"
	"testing"
)

func TestUtils_HashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "success - valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "success - empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "success - long password",
			password: "thisisaverylongpasswordthatshouldstillwork",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if hash != "" {
					t.Errorf("expected empty hash on error, got %s", hash)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if hash == "" {
					t.Error("expected non-empty hash")
				}
				if hash == tt.password {
					t.Error("hash should not equal plain password")
				}
				// Verify hash starts with bcrypt identifier
				if len(hash) < 4 || hash[:4] != "$2a$" {
					t.Errorf("expected bcrypt hash starting with $2a$, got %s", hash[:4])
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// Test with a known password
	password := "testpassword"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "success - correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "failure - wrong password",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "failure - empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "failure - invalid hash",
			password: password,
			hash:     "invalidhash",
			expected: false,
		},
		{
			name:     "failure - empty hash",
			password: password,
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CheckPasswordHash(tt.password, tt.hash)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestPasswordHashingRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"complex!@#$%^&*()",
		"with spaces and symbols !@# 123",
		"unicode: ñáéíóú",
		"verylongpasswordthatexceeds72bytesbutshouldstillworkwithbcrypt",
	}

	for _, password := range passwords {
		t.Run("roundtrip_"+password[:min(10, len(password))], func(t *testing.T) {
			// Hash the password
			hash, err := utils.HashPassword(password)
			if err != nil {
				t.Fatalf("failed to hash password: %v", err)
			}

			// Verify the hash works
			if !utils.CheckPasswordHash(password, hash) {
				t.Error("password hash verification failed")
			}

			// Verify wrong password fails
			if utils.CheckPasswordHash("wrong"+password, hash) {
				t.Error("wrong password should not match hash")
			}
		})
	}
}

func TestPasswordHashingUniqueness(t *testing.T) {
	password := "test-password-123"

	// Generate two hashes for the same password
	hash1, err1 := utils.HashPassword(password)
	if err1 != nil {
		t.Fatalf("failed to hash password first time: %v", err1)
	}

	hash2, err2 := utils.HashPassword(password)
	if err2 != nil {
		t.Fatalf("failed to hash password second time: %v", err2)
	}

	// Hashes should be different due to salt, even for the same password
	if hash1 == hash2 {
		t.Error("expected different hashes for same password due to salt")
	}

	// But both should validate the same password
	if !utils.CheckPasswordHash(password, hash1) {
		t.Error("first hash should validate password")
	}
	if !utils.CheckPasswordHash(password, hash2) {
		t.Error("second hash should validate password")
	}
}

func TestCheckPasswordHash_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "both empty strings",
			password: "",
			hash:     "",
			expected: false,
		},
		{
			name:     "password with only spaces",
			password: "   ",
			hash:     "",
			expected: false,
		},
		{
			name:     "hash with garbage data",
			password: "test",
			hash:     "garbage$hash$data",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CheckPasswordHash(tt.password, tt.hash)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
