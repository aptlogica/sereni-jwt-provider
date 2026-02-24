package models

import "github.com/golang-jwt/jwt/v5"

// User represents a user in the system
type User struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	EMAIL_VERIFIED bool     `json:"email_verified"`
	Roles          []string `json:"roles"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	EMAIL_VERIFIED bool     `json:"email_verified"`
	Roles          []string `json:"roles"`
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken   string   `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	EMAIL_VERIFIED bool     `json:"email_verified"`
	Roles          []string `json:"roles"`
}

// VerifyTokenRequest represents the token verification request payload
type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"900"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Code    string      `json:"code" example:"SUCCESS"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Code    string      `json:"code" example:"ERROR_CODE"`
	Message string      `json:"message" example:"Error message description"`
	Data    interface{} `json:"data"`
}

// CustomClaims represents JWT claims structure
type CustomClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Roles     string `json:"roles"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenClaims represents JWT claims for Swagger documentation
type TokenClaims struct {
	UserID    string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"user@example.com"`
	Roles     string `json:"roles" example:"user,admin"`
	TokenType string `json:"token_type" example:"access"`
	Issuer    string `json:"iss" example:"auth-service"`
	Subject   string `json:"sub" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresAt int64  `json:"exp" example:"1703001600"`
	IssuedAt  int64  `json:"iat" example:"1703000700"`
	NotBefore int64  `json:"nbf" example:"1703000700"`
}
