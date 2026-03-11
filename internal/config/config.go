// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package config

import (
	"fmt"
	"os"
	"strconv"
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

	fmt.Printf("Config loaded: %+v\n", cfg)

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8081"
	}
	if cfg.ServerHost == "" {
		cfg.ServerHost = "0.0.0.0"
	}
	return cfg
}
