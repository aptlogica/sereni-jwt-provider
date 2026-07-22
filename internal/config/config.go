// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
)

const (
	// MinJWTSecretLength enforces minimum entropy for HS256
	// NIST recommends at least 256 bits for HMAC-SHA256
	MinJWTSecretLength = 32
)

type Config struct {
	JWTSecret            string
	ServerPort           string
	ServerHost           string
	AllowedOrigins       string
	Env                  string
	LogLevel             string
	AccessTokenDuration  int64 // in seconds
	RefreshTokenDuration int64 // in seconds
}

// String implements the fmt.Stringer interface to redact sensitive information
func (c Config) String() string {
	return fmt.Sprintf("Config{Port:%s Env:%s JWTSecret:[REDACTED]}", c.ServerPort, c.Env)
}

// ValidateJWTSecret checks that the JWT secret meets minimum strength requirements
// Returns an error if the secret is too short or too predictable
func ValidateJWTSecret(secret string) error {
	if len(secret) == 0 {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(secret) < MinJWTSecretLength {
		return fmt.Errorf(
			"JWT_SECRET is too weak: minimum length is %d characters (256-bit entropy for HS256), got %d. "+
				"Weak secrets like 'secret' or 'changeme' are vulnerable to offline brute-force attacks. "+
				"Use a cryptographically secure random string, e.g., 'openssl rand -base64 32'",
			MinJWTSecretLength,
			len(secret),
		)
	}

	// Check for obvious weak patterns (all same character, sequential, etc.)
	if isWeakPattern(secret) {
		return fmt.Errorf(
			"JWT_SECRET appears to be weak (predictable pattern). " +
				"Use a cryptographically secure random string, e.g., 'openssl rand -base64 32'",
		)
	}

	return nil
}

// isWeakPattern detects obvious weak patterns in the secret
func isWeakPattern(secret string) bool {
	if hasAllSameCharacters(secret) {
		return true
	}
	return hasLowCharacterVariety(secret)
}

// hasAllSameCharacters checks if all characters in the secret are identical (e.g., "aaaaaaa")
func hasAllSameCharacters(secret string) bool {
	if len(secret) == 0 {
		return false
	}
	firstChar := rune(secret[0])
	for _, ch := range secret {
		if ch != firstChar {
			return false
		}
	}
	return true
}

// hasLowCharacterVariety checks if secret lacks sufficient character class diversity
// Returns true if secret has less than 2 character classes
// (allows flexibility for base64 and hex-encoded secrets)
func hasLowCharacterVariety(secret string) bool {
	varietyCount := countCharacterClasses(secret)
	return varietyCount < 2
}

// countCharacterClasses counts distinct character classes in the secret
// Classes: uppercase letters, lowercase letters, digits, special characters
func countCharacterClasses(secret string) int {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range secret {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}
	return count
}

// parseEnvInt64 parses an int64 from string, returns fallback if error
func parseEnvInt64(val string, fallback int64) (int64, error) {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback, err
	}
	return n, nil
}

func LoadConfig() *Config {
	cfg := &Config{
		JWTSecret:      os.Getenv("JWT_SECRET"),
		ServerPort:     os.Getenv("SERVER_PORT"),
		ServerHost:     os.Getenv("SERVER_HOST"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		Env:            os.Getenv("ENV"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
	}
	// Parse AccessTokenDuration and RefreshTokenDuration from env (in seconds)
	if v := os.Getenv("ACCESS_TOKEN_DURATION"); v != "" {
		// fallback to 900 (15min) if parse fails
		if sec, err := parseEnvInt64(v, 900); err == nil {
			cfg.AccessTokenDuration = sec
		} else {
			cfg.AccessTokenDuration = 900
		}
	} else {
		cfg.AccessTokenDuration = 900
	}
	if v := os.Getenv("REFRESH_TOKEN_DURATION"); v != "" {
		// fallback to 604800 (7d) if parse fails
		if sec, err := parseEnvInt64(v, 604800); err == nil {
			cfg.RefreshTokenDuration = sec
		} else {
			cfg.RefreshTokenDuration = 604800
		}
	} else {
		cfg.RefreshTokenDuration = 604800
	}

	log.Printf("Config loaded: %v\n", cfg)

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8081"
	}
	if cfg.ServerHost == "" {
		cfg.ServerHost = "0.0.0.0"
	}
	return cfg
}
