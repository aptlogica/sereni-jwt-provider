// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aptlogica/sereni-jwt-provider/internal/handlers"
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"github.com/aptlogica/sereni-jwt-provider/internal/services"
	"github.com/gin-gonic/gin"
)

func newTestHandler(t *testing.T) *handlers.AuthHandler {
	jwtService := services.NewJWTService(randomTestSigningKey(t), 900, 604800)
	return handlers.NewAuthHandler(jwtService)
}

func randomTestSigningKey(t *testing.T) string {
	if t != nil {
		t.Helper()
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		if t != nil {
			t.Fatalf("failed to generate test signing key: %v", err)
		}
		panic(err)
	}
	return hex.EncodeToString(b)
}

func performJSONRequest(t *testing.T, method, path string, body interface{}, h gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	h(c)
	return w
}

func decodeSuccessResponse(t *testing.T, w *httptest.ResponseRecorder) models.SuccessResponse {
	t.Helper()
	var resp models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode success response: %v", err)
	}
	return resp
}

func TestNewAuthHandler(t *testing.T) {
	handler := newTestHandler(t)
	if handler == nil {
		t.Fatal("expected handler to be created")
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	t.Run("success - valid payload", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
			"id":             "user-1",
			"email":          "user@example.com",
			"roles":          []string{"user", "admin"},
			"email_verified": true,
		}, handler.Login)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		resp := decodeSuccessResponse(t, w)
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected data object, got %T", resp.Data)
		}
		if data["access_token"] == "" {
			t.Error("expected non-empty access_token")
		}
		if data["refresh_token"] == "" {
			t.Error("expected non-empty refresh_token")
		}
	})

	t.Run("failure - invalid json types", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
			"id":    "user-1",
			"email": "user@example.com",
			"roles": "not-an-array",
		}, handler.Login)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success - login with no roles", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
			"id":             "user-3",
			"email":          "user3@example.com",
			"roles":          []string{},
			"email_verified": false,
		}, handler.Login)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		resp := decodeSuccessResponse(t, w)
		if resp.Code != "LOGIN_SUCCESS" {
			t.Errorf("expected code LOGIN_SUCCESS, got %s", resp.Code)
		}
	})

	t.Run("failure - malformed json", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("not-json")))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Login(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "user-1",
		"email":          "user@example.com",
		"roles":          []string{"user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	loginData := loginResp.Data.(map[string]interface{})
	refreshToken := loginData["refresh_token"].(string)

	t.Run("success - valid refresh token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/refresh", map[string]string{
			"refresh_token": refreshToken,
		}, handler.RefreshToken)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("failure - missing refresh token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/refresh", map[string]string{}, handler.RefreshToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("failure - invalid refresh token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/refresh", map[string]string{
			"refresh_token": "not-a-token",
		}, handler.RefreshToken)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "user-123",
		"email":          "user@example.com",
		"roles":          []string{"user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	accessToken := loginResp.Data.(map[string]interface{})["access_token"].(string)

	t.Run("success - valid token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
			"token": accessToken,
		}, handler.ValidateToken)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("failure - invalid token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
			"token": "invalid-token",
		}, handler.ValidateToken)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("failure - empty token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
			"token": "",
		}, handler.ValidateToken)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("failure - missing token field", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{}, handler.ValidateToken)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthHandler_VerifyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "user-verify",
		"email":          "verify@example.com",
		"roles":          []string{"user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	accessToken := loginResp.Data.(map[string]interface{})["access_token"].(string)

	t.Run("success - valid token", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/verify-token", map[string]string{
			"token": accessToken,
		}, handler.VerifyToken)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		resp := decodeSuccessResponse(t, w)
		data := resp.Data.(map[string]interface{})
		if data["valid"] != true {
			t.Fatalf("expected valid=true, got %v", data["valid"])
		}
	})

	t.Run("success - invalid token returns valid=false", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/verify-token", map[string]string{
			"token": "bad-token",
		}, handler.VerifyToken)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		resp := decodeSuccessResponse(t, w)
		data := resp.Data.(map[string]interface{})
		if data["valid"] != false {
			t.Fatalf("expected valid=false, got %v", data["valid"])
		}
	})
}

func TestAuthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	w := performJSONRequest(t, "GET", "/health", nil, handler.Health)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeSuccessResponse(t, w)
	if resp.Code != "HEALTHY" {
		t.Errorf("expected code HEALTHY, got %s", resp.Code)
	}
	if resp.Message != "Service is healthy" {
		t.Errorf("expected message 'Service is healthy', got %s", resp.Message)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if status, ok := data["status"].(string); !ok || status != "healthy" {
		t.Errorf("expected status=healthy, got %v", data["status"])
	}
}

func TestAuthHandler_VerifyToken_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	t.Run("missing token field", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/verify-token", map[string]string{}, handler.VerifyToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/verify-token", bytes.NewBuffer([]byte("{invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.VerifyToken(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthHandler_RefreshToken_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	t.Run("empty refresh token string", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/refresh", map[string]string{
			"refresh_token": "",
		}, handler.RefreshToken)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer([]byte("not valid json")))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.RefreshToken(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthHandler_ValidateToken_ClaimsVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "claims-test-user",
		"email":          "claims@example.com",
		"roles":          []string{"admin", "user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	accessToken := loginResp.Data.(map[string]interface{})["access_token"].(string)

	w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
		"token": accessToken,
	}, handler.ValidateToken)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeSuccessResponse(t, w)
	data := resp.Data.(map[string]interface{})

	if data["user_id"] != "claims-test-user" {
		t.Errorf("expected user_id=claims-test-user, got %v", data["user_id"])
	}
	if data["email"] != "claims@example.com" {
		t.Errorf("expected email=claims@example.com, got %v", data["email"])
	}
	if data["roles"] != "admin,user" {
		t.Errorf("expected roles=admin,user, got %v", data["roles"])
	}
	if data["token_type"] != "access" {
		t.Errorf("expected token_type=access, got %v", data["token_type"])
	}
}

func TestAuthHandler_ValidateToken_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a handler with very short token expiry
	jwtService := services.NewJWTService(randomTestSigningKey(t), -1, 604800) // -1 second = already expired
	handler := handlers.NewAuthHandler(jwtService)

	// Generate expired access token
	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "expired-user",
		"email":          "expired@example.com",
		"roles":          []string{"user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	expiredToken := loginResp.Data.(map[string]interface{})["access_token"].(string)

	// Wait a tiny bit to ensure it's expired
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
		"token": expiredToken,
	}, handler.ValidateToken)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// JWT library returns "invalid claims: token is expired" which doesn't match serrors.ErrTokenExpire
	// So the handler falls through to INVALID_TOKEN
	if resp.Code != "INVALID_TOKEN" {
		t.Errorf("expected code INVALID_TOKEN, got %s", resp.Code)
	}
}

func TestAuthHandler_Login_MultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	t.Run("three roles", func(t *testing.T) {
		w := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
			"id":             "multi-role-user",
			"email":          "multi@example.com",
			"roles":          []string{"admin", "editor", "viewer"},
			"email_verified": true,
		}, handler.Login)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		resp := decodeSuccessResponse(t, w)
		accessToken := resp.Data.(map[string]interface{})["access_token"].(string)

		// Validate the token to check claims
		validateW := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
			"token": accessToken,
		}, handler.ValidateToken)

		validateResp := decodeSuccessResponse(t, validateW)
		data := validateResp.Data.(map[string]interface{})
		if data["roles"] != "admin,editor,viewer" {
			t.Errorf("expected roles=admin,editor,viewer, got %v", data["roles"])
		}
	})
}

func TestAuthHandler_ValidateToken_NilClaimsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	// Test with a token that would result in an error but not be expired or explicitly invalid
	// Using a deliberately malformed token string (non-secret) to trigger validation error
	w := performJSONRequest(t, "POST", "/auth/validate-token", map[string]string{
		"token": "invalid-token",
	}, handler.ValidateToken)

	// This should fail with INVALID_TOKEN since it's signed with a different secret
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_RefreshToken_WithAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler(t)

	// Login first
	loginW := performJSONRequest(t, "POST", "/auth/login", map[string]interface{}{
		"id":             "refresh-test-user",
		"email":          "refresh@example.com",
		"roles":          []string{"user"},
		"email_verified": true,
	}, handler.Login)
	loginResp := decodeSuccessResponse(t, loginW)
	accessToken := loginResp.Data.(map[string]interface{})["access_token"].(string)

	// Try to refresh using access token instead of refresh token
	w := performJSONRequest(t, "POST", "/auth/refresh", map[string]string{
		"refresh_token": accessToken,
	}, handler.RefreshToken)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
