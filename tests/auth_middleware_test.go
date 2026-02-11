package tests

import (
	"auth-service/internal/middleware"
	middlewarePkg "auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"auth-service/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		setupToken     bool
		tokenType      string
		expectAbort    bool
	}{
		{
			name:           "success - valid access token",
			authHeader:     "Bearer valid-token",
			expectedStatus: http.StatusOK,
			setupToken:     true,
			tokenType:      "access",
			expectAbort:    false,
		},
		{
			name:           "failure - missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     false,
			expectAbort:    true,
		},
		{
			name:           "failure - invalid header format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     false,
			expectAbort:    true,
		},
		{
			name:           "failure - invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     false,
			expectAbort:    true,
		},
		{
			name:           "failure - refresh token instead of access",
			authHeader:     "Bearer refresh-token",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     true,
			tokenType:      "refresh",
			expectAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)

			// Setup user and token if needed
			var token string
			var userID string
			if tt.setupToken {
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test-user-id", "test@example.com", hashedPassword, []string{"user"})
				userID = user.ID

				pair, _ := jwtService.GenerateTokenPair(user)
				if tt.tokenType == "refresh" {
					token = pair.RefreshToken
				} else {
					token = pair.AccessToken
				}
			}

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set auth header
			if tt.authHeader != "" {
				if tt.authHeader == "Bearer valid-token" && tt.setupToken {
					c.Request = httptest.NewRequest("GET", "/", nil)
					c.Request.Header.Set("Authorization", "Bearer "+token)
				} else {
					c.Request = httptest.NewRequest("GET", "/", nil)
					c.Request.Header.Set("Authorization", tt.authHeader)
				}
			} else {
				c.Request = httptest.NewRequest("GET", "/", nil)
			}

			// Create middleware
			middleware := middleware.AuthMiddleware(jwtService)

			// Call middleware
			middleware(c)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check if aborted
			if tt.expectAbort && !c.IsAborted() {
				t.Error("expected request to be aborted")
			}

			// Check context values for success case
			if tt.name == "success - valid access token" && !c.IsAborted() {
				ctxUserID, exists := c.Get("userID")
				if !exists || ctxUserID != userID {
					t.Errorf("expected userID %s in context, got %v", userID, ctxUserID)
				}
				email, exists := c.Get("email")
				if !exists || email != "test@example.com" {
					t.Error("expected email in context")
				}
				roles, exists := c.Get("roles")
				if !exists {
					t.Error("expected roles in context")
				}
				userRoles, ok := roles.([]string)
				if !ok || len(userRoles) != 1 || userRoles[0] != "user" {
					t.Error("expected correct roles in context")
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRole   string
		contextRoles   interface{}
		expectedStatus int
		expectAbort    bool
	}{
		{
			name:           "success - user has required role",
			requiredRole:   "admin",
			contextRoles:   []string{"user", "admin"},
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "failure - user does not have required role",
			requiredRole:   "admin",
			contextRoles:   []string{"user"},
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "failure - no roles in context",
			requiredRole:   "admin",
			contextRoles:   nil,
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "failure - invalid roles format",
			requiredRole:   "admin",
			contextRoles:   "invalid",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			// Set roles in context if provided
			if tt.contextRoles != nil {
				c.Set("roles", tt.contextRoles)
			}

			// Create middleware
			middleware := middlewarePkg.RequireRole(tt.requiredRole)

			// Call middleware
			middleware(c)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check if aborted
			if tt.expectAbort && !c.IsAborted() {
				t.Error("expected request to be aborted")
			}
		})
	}
}
