package services

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	TokenTypeAccess      = "access"
	TokenTypeRefresh     = "refresh"
	Issuer               = "auth-service"
)

// JWTService handles JWT operations
type JWTService struct {
	secretKey  string
	tokenStore *repository.TokenStore
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, tokenStore *repository.TokenStore) *JWTService {
	return &JWTService{
		secretKey:  secretKey,
		tokenStore: tokenStore,
	}
}

// Register registers a new user
func (s *JWTService) Register(email, password string, roles []string) (*models.User, error) {
	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.tokenStore.CreateUser(email, hashedPassword, roles)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *JWTService) Login(email, password string) (*models.TokenResponse, error) {
	// Get user
	user, err := s.tokenStore.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	return s.GenerateTokenPair(user)
}

// GenerateTokenPair generates access and refresh tokens
func (s *JWTService) GenerateTokenPair(user *models.User) (*models.TokenResponse, error) {
	// Generate access token
	accessToken, err := s.generateToken(user, TokenTypeAccess, AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(user, TokenTypeRefresh, RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.tokenStore.StoreRefreshToken(user.ID, refreshToken); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
	}, nil
}

// generateToken generates a JWT token
func (s *JWTService) generateToken(user *models.User, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := models.CustomClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Roles:     user.Roles,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*models.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken refreshes the access token using a refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string) (*models.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check token type
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("token is not a refresh token")
	}

	// Validate refresh token in store
	userID, err := s.tokenStore.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if userID != claims.UserID {
		return nil, errors.New("token user mismatch")
	}

	// Get user
	user, err := s.tokenStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	s.tokenStore.RevokeRefreshToken(refreshToken)

	// Generate new token pair (token rotation)
	return s.GenerateTokenPair(user)
}

// Logout revokes the refresh token
func (s *JWTService) Logout(refreshToken string) error {
	return s.tokenStore.RevokeRefreshToken(refreshToken)
}

// LogoutAll revokes all refresh tokens for a user
func (s *JWTService) LogoutAll(userID string) error {
	return s.tokenStore.RevokeAllUserTokens(userID)
}

// GetUserProfile retrieves user profile information
func (s *JWTService) GetUserProfile(userID string) (*models.UserProfile, error) {
	user, err := s.tokenStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &models.UserProfile{
		ID:    user.ID,
		Email: user.Email,
		Roles: user.Roles,
	}, nil
}
