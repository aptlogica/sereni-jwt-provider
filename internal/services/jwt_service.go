package services

import (
	serrors "auth-service/internal/errors"
	"auth-service/internal/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	Issuer           = "auth-service"
)

// JWTService handles JWT operations
type JWTService struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenDuration, refreshTokenDuration int64) *JWTService {
	svc := &JWTService{
		secretKey:            secretKey,
		accessTokenDuration:  time.Duration(accessTokenDuration) * time.Second,
		refreshTokenDuration: time.Duration(refreshTokenDuration) * time.Second,
	}
	fmt.Printf("[DEBUG] JWTService: accessTokenDuration=%v seconds, refreshTokenDuration=%v seconds\n", svc.accessTokenDuration.Seconds(), svc.refreshTokenDuration.Seconds())
	return svc
}

// Login authenticates a user and returns tokens
func (s *JWTService) Login(loginReq *models.LoginRequest) (*models.TokenResponse, error) {
	user := &models.User{
		ID:             loginReq.ID,
		Email:          loginReq.Email,
		Roles:          loginReq.Roles,
		EMAIL_VERIFIED: loginReq.EMAIL_VERIFIED,
	}

	// Generate tokens
	return s.GenerateTokenPair(user)
}

// GenerateTokenPair generates access and refresh tokens
func (s *JWTService) GenerateTokenPair(user *models.User) (*models.TokenResponse, error) {
	fmt.Printf("[DEBUG] GenerateTokenPair: userID=%s, email=%s, roles=%v\n", user.ID, user.Email, user.Roles, user.EMAIL_VERIFIED)
	// Generate access token
	accessToken, err := s.generateToken(user, TokenTypeAccess, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(user, TokenTypeRefresh, s.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	// Store refresh token for revocation
	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenDuration.Seconds()),
	}, nil
}

// generateToken generates a JWT token
func (s *JWTService) generateToken(user *models.User, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	exp := now.Add(duration)
	fmt.Printf("[DEBUG] generateToken: now=%v, duration=%v, exp=%v (seconds diff=%v)\n", now.Unix(), duration.Seconds(), exp.Unix(), exp.Unix()-now.Unix())

	// Convert roles array to comma-separated string
	rolesStr := ""
	if len(user.Roles) > 0 {
		rolesStr = user.Roles[0]
		for i := 1; i < len(user.Roles); i++ {
			rolesStr += "," + user.Roles[i]
		}
	}

	claims := models.CustomClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Roles:     rolesStr,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(exp),
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
			return nil, serrors.ErrUnexpectedSigningMethod
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, serrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.CustomClaims)
	if !ok {
		return nil, serrors.ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, serrors.ErrTokenExpire // or a more specific error for expired token
	}

	return claims, nil
}

// RefreshAccessToken refreshes the access token using a refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string) (*models.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, serrors.ErrInvalidRefreshTokenSvc
	}

	// Check token type
	if claims.TokenType != TokenTypeRefresh {
		return nil, serrors.ErrNotRefreshToken
	}

	user := &models.User{
		ID:             claims.UserID,
		Email:          claims.Email,
		EMAIL_VERIFIED: claims.EMAIL_VERIFIED,
		Roles:          []string{},
	}

	// Generate new token pair (token rotation)
	return s.GenerateTokenPair(user)
}
