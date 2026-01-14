package repository

import (
	serrors "auth-service/internal/errors"
	"auth-service/internal/models"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RefreshMeta stores metadata about a refresh token
type RefreshMeta struct {
	UserID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// TokenStore handles in-memory storage of users and refresh tokens
type TokenStore struct {
	users         map[string]*models.User // email -> user
	refreshTokens map[string]RefreshMeta  // hashedRefreshToken -> meta
	userTokens    map[string][]string     // userID -> []hashedRefreshTokens
	mu            sync.RWMutex
	stop          chan struct{}
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	return &TokenStore{
		users:         make(map[string]*models.User),
		refreshTokens: make(map[string]RefreshMeta),
		userTokens:    make(map[string][]string),
	}
}

// CreateUser creates a new user
func (ts *TokenStore) CreateUser(email, hashedPassword string, roles []string) (*models.User, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.users[email]; exists {
		return nil, serrors.ErrUserExists
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
		return nil, serrors.ErrUserNotFound
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

	return nil, serrors.ErrUserNotFound
}

// StoreRefreshToken stores a refresh token (stores only a hash + metadata)
func (ts *TokenStore) StoreRefreshToken(userID, refreshToken string, expiresAt time.Time) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Hash the refresh token before storing
	h := sha256.Sum256([]byte(refreshToken))
	key := hex.EncodeToString(h[:])

	ts.refreshTokens[key] = RefreshMeta{
		UserID:    userID,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: expiresAt,
	}

	ts.userTokens[userID] = append(ts.userTokens[userID], key)

	// Clean up old tokens (keep last 5 per user)
	tokens := ts.userTokens[userID]
	if len(tokens) > 5 {
		oldKey := tokens[0]
		delete(ts.refreshTokens, oldKey)
		ts.userTokens[userID] = tokens[1:]
	}

	return nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (ts *TokenStore) ValidateRefreshToken(refreshToken string) (string, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	h := sha256.Sum256([]byte(refreshToken))
	key := hex.EncodeToString(h[:])

	meta, exists := ts.refreshTokens[key]
	if !exists {
		return "", serrors.ErrInvalidRefreshToken
	}

	// Check expiry
	if time.Now().UTC().After(meta.ExpiresAt) {
		return "", serrors.ErrInvalidRefreshToken
	}

	return meta.UserID, nil
}

// RevokeRefreshToken revokes a refresh token
func (ts *TokenStore) RevokeRefreshToken(refreshToken string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	h := sha256.Sum256([]byte(refreshToken))
	key := hex.EncodeToString(h[:])

	meta, exists := ts.refreshTokens[key]
	if !exists {
		return serrors.ErrTokenNotFound
	}

	delete(ts.refreshTokens, key)

	// Remove from user's token list
	tokens := ts.userTokens[meta.UserID]
	for i, tokenKey := range tokens {
		if tokenKey == key {
			ts.userTokens[meta.UserID] = append(tokens[:i], tokens[i+1:]...)
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

	for _, tokenKey := range tokens {
		delete(ts.refreshTokens, tokenKey)
	}

	delete(ts.userTokens, userID)
	return nil
}

// CleanExpiredTokens removes expired tokens (this should be called periodically)
func (ts *TokenStore) CleanExpiredTokens() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now().UTC()
	for key, meta := range ts.refreshTokens {
		if now.After(meta.ExpiresAt) {
			delete(ts.refreshTokens, key)
			// remove from user's list
			tokens := ts.userTokens[meta.UserID]
			for i := 0; i < len(tokens); i++ {
				if tokens[i] == key {
					ts.userTokens[meta.UserID] = append(tokens[:i], tokens[i+1:]...)
					break
				}
			}
		}
	}
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
	if ts.stop != nil {
		return // already started
	}
	ts.stop = make(chan struct{})
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ts.CleanExpiredTokens()
			case <-ts.stop:
				return
			}
		}
	}()
}

// Close stops background routines for the token store
func (ts *TokenStore) Close() {
	if ts.stop == nil {
		return
	}
	close(ts.stop)
	ts.stop = nil
}
