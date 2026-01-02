package repository

import (
	"auth-service/internal/models"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TokenStore handles in-memory storage of users and refresh tokens
type TokenStore struct {
	users         map[string]*models.User // email -> user
	refreshTokens map[string]string       // refreshToken -> userID
	userTokens    map[string][]string     // userID -> []refreshTokens
	mu            sync.RWMutex
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	return &TokenStore{
		users:         make(map[string]*models.User),
		refreshTokens: make(map[string]string),
		userTokens:    make(map[string][]string),
	}
}

// CreateUser creates a new user
func (ts *TokenStore) CreateUser(email, hashedPassword string, roles []string) (*models.User, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.users[email]; exists {
		return nil, errors.New("user already exists")
	}

	if roles == nil || len(roles) == 0 {
		roles = []string{"user"}
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Password: hashedPassword,
		Roles:    roles,
	}

	ts.users[email] = user
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (ts *TokenStore) GetUserByEmail(email string) (*models.User, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	user, exists := ts.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (ts *TokenStore) GetUserByID(userID string) (*models.User, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, user := range ts.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// StoreRefreshToken stores a refresh token
func (ts *TokenStore) StoreRefreshToken(userID, refreshToken string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.refreshTokens[refreshToken] = userID
	ts.userTokens[userID] = append(ts.userTokens[userID], refreshToken)

	// Clean up old tokens (keep last 5 per user)
	tokens := ts.userTokens[userID]
	if len(tokens) > 5 {
		oldToken := tokens[0]
		delete(ts.refreshTokens, oldToken)
		ts.userTokens[userID] = tokens[1:]
	}

	return nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (ts *TokenStore) ValidateRefreshToken(refreshToken string) (string, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	userID, exists := ts.refreshTokens[refreshToken]
	if !exists {
		return "", errors.New("invalid refresh token")
	}

	return userID, nil
}

// RevokeRefreshToken revokes a refresh token
func (ts *TokenStore) RevokeRefreshToken(refreshToken string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	userID, exists := ts.refreshTokens[refreshToken]
	if !exists {
		return errors.New("token not found")
	}

	delete(ts.refreshTokens, refreshToken)

	// Remove from user's token list
	tokens := ts.userTokens[userID]
	for i, token := range tokens {
		if token == refreshToken {
			ts.userTokens[userID] = append(tokens[:i], tokens[i+1:]...)
			break
		}
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (ts *TokenStore) RevokeAllUserTokens(userID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tokens, exists := ts.userTokens[userID]
	if !exists {
		return nil
	}

	for _, token := range tokens {
		delete(ts.refreshTokens, token)
	}

	delete(ts.userTokens, userID)
	return nil
}

// CleanExpiredTokens removes expired tokens (this should be called periodically)
func (ts *TokenStore) CleanExpiredTokens() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// In a real implementation, you would track token expiration times
	// For now, this is a placeholder for periodic cleanup
	// You might implement this with a time.Ticker in production
}

// GetStats returns storage statistics
func (ts *TokenStore) GetStats() map[string]int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return map[string]int{
		"users":          len(ts.users),
		"refresh_tokens": len(ts.refreshTokens),
	}
}

// StartCleanupRoutine starts a goroutine to periodically clean up expired tokens
func (ts *TokenStore) StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			ts.CleanExpiredTokens()
		}
	}()
}
