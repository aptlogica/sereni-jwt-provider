package services

import (
	serrors "auth-service/internal/errors"
	"auth-service/internal/models"
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
	secretKey string
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: secretKey,
	}
}

// Login authenticates a user and returns tokens
func (s *JWTService) Login(loginReq *models.LoginRequest) (*models.TokenResponse, error) {
	user := &models.User{
		ID:       loginReq.ID,
		Email:    loginReq.Email,
		Password: loginReq.Password,
		Roles:    loginReq.Roles,
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

	// Store refresh token for revocation
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
		return nil, serrors.ErrInvalidToken // or a more specific error for expired token
	}

	return claims, nil
}

// RefreshAccessToken refreshes the access token using a refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string, user *models.User) (*models.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, serrors.ErrInvalidRefreshTokenSvc
	}

	// Check token type
	if claims.TokenType != TokenTypeRefresh {
		return nil, serrors.ErrNotRefreshToken
	}

	// Generate new token pair (token rotation)
	return s.GenerateTokenPair(user)
}
