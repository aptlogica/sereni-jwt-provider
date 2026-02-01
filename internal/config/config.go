package config

import (
	"os"
)

type Config struct {
	JWTSecret      string
	ServerPort     string
	ServerHost     string
	AllowedOrigins string
	Env            string
	LogLevel       string
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
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8081"
	}
	if cfg.ServerHost == "" {
		cfg.ServerHost = "0.0.0.0"
	}
	return cfg
}
