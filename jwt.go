// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

// Package sereni_jwt_provider provides a simple JWT authentication library.
// This is the public API for the Sereni JWT Provider.
package sereni_jwt_provider

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds the JWT provider configuration.
type Config struct {
	// Secret is the key used to sign and verify tokens
	Secret string
	// Expiry is the duration until the token expires
	Expiry time.Duration
	// Issuer is the issuer claim for the token
	Issuer string
	// Algorithm is the signing algorithm (e.g., "HS256", "HS384", "HS512")
	Algorithm string
}

// Provider is the JWT provider that handles token generation and validation.
type Provider struct {
	config Config
}

// NewProvider creates a new JWT provider with the given configuration.
func NewProvider(config Config) *Provider {
	return &Provider{config: config}
}

// GenerateToken creates a new JWT token with the given claims.
func (p *Provider) GenerateToken(claims map[string]interface{}) (string, error) {
	if p.config.Secret == "" {
		return "", errors.New("secret key is required")
	}

	now := time.Now()

	// Create JWT claims
	jwtClaims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(p.config.Expiry).Unix(),
		"iss": p.config.Issuer,
	}

	// Add custom claims
	for key, value := range claims {
		jwtClaims[key] = value
	}

	// Determine signing method
	var signingMethod jwt.SigningMethod
	switch p.config.Algorithm {
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	default:
		signingMethod = jwt.SigningMethodHS256
	}

	// Create and sign the token
	token := jwt.NewWithClaims(signingMethod, jwtClaims)
	signedToken, err := token.SignedString([]byte(p.config.Secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (p *Provider) ValidateToken(tokenString string) (map[string]interface{}, error) {
	if p.config.Secret == "" {
		return nil, errors.New("secret key is required")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		switch p.config.Algorithm {
		case "HS384":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
		case "HS512":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
		default:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
		}
		return []byte(p.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// Convert to map[string]interface{}
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// RefreshToken generates a new token from an existing valid token.
func (p *Provider) RefreshToken(tokenString string) (string, error) {
	claims, err := p.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Remove standard claims that will be regenerated
	delete(claims, "iat")
	delete(claims, "exp")
	delete(claims, "iss")

	return p.GenerateToken(claims)
}
