package tests

import (
	"auth-service/internal/handlers"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"auth-service/internal/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewAuthHandler(t *testing.T) {
	tokenStore := repository.NewTokenStore()
	jwtService := services.NewJWTService("test-secret", tokenStore)
	handler := handlers.NewAuthHandler(jwtService)

	if handler == nil {
		t.Errorf("expected handler to be created")
	}
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupUser      bool
		checkResponse  bool
	}{
		{
			name: "success - valid registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
				"roles":    []string{"user"},
			},
			expectedStatus: http.StatusCreated,
			setupUser:      false,
			checkResponse:  true,
		},
		{
			name: "failure - invalid JSON (missing password)",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			setupUser:      false,
			checkResponse:  false,
		},
		{
			name: "failure - empty email",
			requestBody: map[string]interface{}{
				"email":    "",
				"password": "password123",
				"roles":    []string{"user"},
			},
			expectedStatus: http.StatusBadRequest,
			setupUser:      false,
			checkResponse:  false,
		},
		{
			name: "failure - user already exists",
			requestBody: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "password123",
				"roles":    []string{"user"},
			},
			expectedStatus: http.StatusConflict,
			setupUser:      true,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			// Setup existing user if needed
			if tt.setupUser {
				hashedPassword, _ := utils.HashPassword("password123")
				tokenStore.CreateUser("existing@example.com", hashedPassword, []string{"user"})
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Register(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Verify response structure on success
			if tt.checkResponse && w.Code == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if data, ok := response["data"].(map[string]interface{}); ok {
					if email, hasEmail := data["email"]; !hasEmail || email != tt.requestBody.(map[string]interface{})["email"] {
						t.Error("expected email in response data")
					}
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupUser      bool
		checkResponse  bool
	}{
		{
			name: "success - valid login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			setupUser:      true,
			checkResponse:  true,
		},
		{
			name: "failure - invalid JSON",
			requestBody: map[string]interface{}{
				"email": "",
			},
			expectedStatus: http.StatusBadRequest,
			setupUser:      false,
			checkResponse:  false,
		},
		{
			name: "failure - empty email",
			requestBody: map[string]string{
				"email":    "",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			setupUser:      false,
			checkResponse:  false,
		},
		{
			name: "failure - user not found",
			requestBody: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			setupUser:      false,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			// Setup user if needed
			if tt.setupUser {
				hashedPassword, _ := utils.HashPassword("password123")
				tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Login(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check that tokens are returned on success
			if tt.checkResponse && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if data, ok := response["data"].(map[string]interface{}); ok {
					if _, hasAccess := data["access_token"]; !hasAccess {
						t.Error("expected access_token in response")
					}
					if _, hasRefresh := data["refresh_token"]; !hasRefresh {
						t.Error("expected refresh_token in response")
					}
				}
			}
		})
	}
}

func TestAuthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenStore := repository.NewTokenStore()
	jwtService := services.NewJWTService("test-secret", tokenStore)
	handler := handlers.NewAuthHandler(jwtService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	handler.Health(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupTokens    bool
	}{
		{
			name: "success - valid refresh token",
			requestBody: map[string]string{
				"refresh_token": "valid-token",
			},
			expectedStatus: http.StatusOK,
			setupTokens:    true,
		},
		{
			name: "failure - invalid JSON",
			requestBody: map[string]interface{}{
				"refresh_token": 123, // wrong type
			},
			expectedStatus: http.StatusBadRequest,
			setupTokens:    false,
		},
		{
			name: "failure - invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid-token",
			},
			expectedStatus: http.StatusUnauthorized,
			setupTokens:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			var refreshToken string
			if tt.setupTokens {
				// Setup user and get valid refresh token
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
				pair, _ := jwtService.GenerateTokenPair(user)
				refreshToken = pair.RefreshToken
				tt.requestBody.(map[string]string)["refresh_token"] = refreshToken
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.RefreshToken(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupTokens    bool
	}{
		{
			name: "success - valid logout",
			requestBody: map[string]string{
				"refresh_token": "valid-token",
			},
			expectedStatus: http.StatusOK,
			setupTokens:    true,
		},
		{
			name: "failure - invalid JSON",
			requestBody: map[string]interface{}{
				"refresh_token": nil, // invalid
			},
			expectedStatus: http.StatusBadRequest,
			setupTokens:    false,
		},
		{
			name: "failure - invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid-token",
			},
			expectedStatus: http.StatusBadRequest,
			setupTokens:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			var refreshToken string
			if tt.setupTokens {
				// Setup user and get valid refresh token
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
				pair, _ := jwtService.GenerateTokenPair(user)
				refreshToken = pair.RefreshToken
				tt.requestBody.(map[string]string)["refresh_token"] = refreshToken
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Logout(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupTokens    bool
	}{
		{
			name: "success - valid token",
			requestBody: map[string]string{
				"token": "valid-token",
			},
			expectedStatus: http.StatusOK,
			setupTokens:    true,
		},
		{
			name: "failure - invalid JSON",
			requestBody: map[string]interface{}{
				"token": 123, // wrong type
			},
			expectedStatus: http.StatusBadRequest,
			setupTokens:    false,
		},
		{
			name: "failure - invalid token",
			requestBody: map[string]string{
				"token": "invalid-token",
			},
			expectedStatus: http.StatusUnauthorized,
			setupTokens:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			var accessToken string
			if tt.setupTokens {
				// Setup user and get valid access token
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
				pair, _ := jwtService.GenerateTokenPair(user)
				accessToken = pair.AccessToken
				tt.requestBody.(map[string]string)["token"] = accessToken
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/validate-token", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ValidateToken(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_VerifyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupTokens    bool
		expectValid    bool
	}{
		{
			name: "success - valid token",
			requestBody: map[string]string{
				"token": "valid-token",
			},
			expectedStatus: http.StatusOK,
			setupTokens:    true,
			expectValid:    true,
		},
		{
			name: "success - invalid token",
			requestBody: map[string]string{
				"token": "invalid-token",
			},
			expectedStatus: http.StatusOK,
			setupTokens:    false,
			expectValid:    false,
		},
		{
			name: "failure - invalid JSON",
			requestBody: map[string]interface{}{
				"token": 123, // wrong type
			},
			expectedStatus: http.StatusBadRequest,
			setupTokens:    false,
			expectValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			var accessToken string
			if tt.setupTokens {
				// Setup user and get valid access token
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
				pair, _ := jwtService.GenerateTokenPair(user)
				accessToken = pair.AccessToken
				tt.requestBody.(map[string]string)["token"] = accessToken
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/verify-token", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.VerifyToken(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful requests, check the valid field
			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data := response["data"].(map[string]interface{})
				valid := data["valid"].(bool)
				if valid != tt.expectValid {
					t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, valid)
				}
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupUserID    bool
		userID         string
		expectedStatus int
	}{
		{
			name:           "success - valid profile request",
			setupUserID:    true,
			userID:         "valid-user-id",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failure - no userID in context",
			setupUserID:    false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "failure - user not found",
			setupUserID:    true,
			userID:         "nonexistent-user",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStore := repository.NewTokenStore()
			jwtService := services.NewJWTService("test-secret", tokenStore)
			handler := handlers.NewAuthHandler(jwtService)

			var userID string
			if tt.setupUserID && tt.userID == "valid-user-id" {
				// Setup user
				hashedPassword, _ := utils.HashPassword("password123")
				user, _ := tokenStore.CreateUser("test@example.com", hashedPassword, []string{"user"})
				userID = user.ID
			} else if tt.setupUserID {
				userID = tt.userID
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/profile", nil)

			if tt.setupUserID {
				c.Set("userID", userID)
			}

			handler.GetProfile(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
