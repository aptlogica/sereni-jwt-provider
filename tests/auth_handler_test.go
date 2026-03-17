// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aptlogica/sereni-jwt-provider/internal/handlers"
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"github.com/aptlogica/sereni-jwt-provider/internal/services"
	"github.com/gin-gonic/gin"
)

func newTestHandler() *handlers.AuthHandler {
	jwtService := services.NewJWTService("test-secret", 900, 604800)
	return handlers.NewAuthHandler(jwtService)
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
	handler := newTestHandler()
	if handler == nil {
		t.Fatal("expected handler to be created")
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
	handler := newTestHandler()

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
