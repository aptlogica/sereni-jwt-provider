package tests

import (
	"auth-service/internal/middleware"
	"auth-service/internal/models"
	"auth-service/internal/services"
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
			tokenType:      services.TokenTypeAccess,
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
			tokenType:      services.TokenTypeRefresh,
			expectAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := services.NewJWTService("test-secret")

			var token string
			user := &models.User{
				ID:       "test-user-id",
				Email:    "test@example.com",
				Password: "password123",
				Roles:    []string{"user", "admin"},
			}

			if tt.setupToken {
				pair, err := jwtService.GenerateTokenPair(user)
				if err != nil {
					t.Fatalf("failed generating token pair: %v", err)
				}
				if tt.tokenType == services.TokenTypeRefresh {
					token = pair.RefreshToken
				} else {
					token = pair.AccessToken
				}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			if tt.authHeader != "" {
				if tt.authHeader == "Bearer valid-token" && tt.setupToken {
					c.Request.Header.Set("Authorization", "Bearer "+token)
				} else if tt.authHeader == "Bearer refresh-token" && tt.setupToken {
					c.Request.Header.Set("Authorization", "Bearer "+token)
				} else {
					c.Request.Header.Set("Authorization", tt.authHeader)
				}
			}

			handler := middleware.AuthMiddleware(jwtService)
			handler(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectAbort && !c.IsAborted() {
				t.Error("expected request to be aborted")
			}

			if tt.name == "success - valid access token" && !c.IsAborted() {
				ctxUserID, exists := c.Get("userID")
				if !exists || ctxUserID != user.ID {
					t.Errorf("expected userID %s in context, got %v", user.ID, ctxUserID)
				}

				email, exists := c.Get("email")
				if !exists || email != user.Email {
					t.Errorf("expected email %s in context, got %v", user.Email, email)
				}

				roles, exists := c.Get("roles")
				if !exists {
					t.Error("expected roles in context")
				}
				userRoles, ok := roles.([]string)
				if !ok || len(userRoles) != 2 || userRoles[0] != "user" || userRoles[1] != "admin" {
					t.Errorf("expected [user admin] roles in context, got %v", roles)
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
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			if tt.contextRoles != nil {
				c.Set("roles", tt.contextRoles)
			}

			handler := middleware.RequireRole(tt.requiredRole)
			handler(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectAbort && !c.IsAborted() {
				t.Error("expected request to be aborted")
			}
		})
	}
}
