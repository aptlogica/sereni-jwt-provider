// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
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

// getTestSecretKey generates a test secret key for JWT operations
func getTestSecretKey() string {
	// Using a base string that meets JWT requirements (32+ bytes)
	return "test" + "secret" + "key" + "must" + "be" + "longer" + "than" + "32" + "bytes"
}

func setupAuthHandler(t *testing.T) *handlers.AuthHandler {
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(15 * 60)        // 15 minutes in seconds
	refreshTokenDuration := int64(7 * 24 * 3600) // 7 days in seconds

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	return handlers.NewAuthHandler(jwtService)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	loginReq := models.LoginRequest{
		ID:             "user123",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	}

	body, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success true, got false")
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	loginReq := models.LoginRequest{
		// Empty request
	}

	body, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	// Current behavior: login does not validate missing fields and returns 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	refreshReq := models.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}

	body, _ := json.Marshal(refreshReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_ValidateToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/validate-token", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateToken(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_ValidateToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	validateReq := models.VerifyTokenRequest{
		Token: "invalid-token",
	}

	body, _ := json.Marshal(validateReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/validate-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_VerifyToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/verify-token", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyToken(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_VerifyToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	verifyReq := models.VerifyTokenRequest{
		Token: "invalid-token",
	}

	body, _ := json.Marshal(verifyReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/verify-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyToken(c)

	// Current behavior: verify returns 200 with valid=false
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_NewAuthHandler(t *testing.T) {
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(15 * 60)
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	handler := handlers.NewAuthHandler(jwtService)

	if handler == nil {
		t.Error("expected handler to not be nil")
	}
}

func TestAuthHandler_VerifyToken_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a token with very short expiry that's already expired
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(-5) // Already expired
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	handler := handlers.NewAuthHandler(jwtService)

	loginReq := models.LoginRequest{
		ID:             "user123",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	}

	tokens, err := jwtService.Login(&loginReq)
	if err != nil {
		t.Fatalf("failed to get tokens: %v", err)
	}

	verifyReq := models.VerifyTokenRequest{
		Token: tokens.AccessToken,
	}

	body, _ := json.Marshal(verifyReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/verify-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyToken(c)

	// VerifyToken always returns 200 with a valid boolean
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_ValidateToken_TokenExpired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a token with very short expiry that's already expired
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(-5) // Already expired
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	handler := handlers.NewAuthHandler(jwtService)

	loginReq := models.LoginRequest{
		ID:             "user123",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	}

	tokens, err := jwtService.Login(&loginReq)
	if err != nil {
		t.Fatalf("failed to get tokens: %v", err)
	}

	validateReq := models.VerifyTokenRequest{
		Token: tokens.AccessToken,
	}

	body, _ := json.Marshal(validateReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/validate-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateToken(c)

	// Should return 401 for expired token
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	handler.Health(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success true, got false")
	}
	if response.Code != "HEALTHY" {
		t.Errorf("expected code HEALTHY, got %s", response.Code)
	}
}

func TestAuthHandler_VerifyToken_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a valid token using the same JWT service to ensure compatibility
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(15 * 60)
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	validHandler := handlers.NewAuthHandler(jwtService)

	loginReq := models.LoginRequest{
		ID:             "user123",
		Email:          "test@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	}

	tokens, err := jwtService.Login(&loginReq)
	if err != nil {
		t.Fatalf("failed to get tokens: %v", err)
	}

	verifyReq := models.VerifyTokenRequest{
		Token: tokens.AccessToken,
	}

	body, _ := json.Marshal(verifyReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/verify-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	validHandler.VerifyToken(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_ValidateToken_WithClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use dedicated handler and service for this test
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(15 * 60)
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	handler := handlers.NewAuthHandler(jwtService)

	loginReq := models.LoginRequest{
		ID:             "user456",
		Email:          "admin@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"admin", "user"},
	}

	tokens, err := jwtService.Login(&loginReq)
	if err != nil {
		t.Fatalf("failed to get tokens: %v", err)
	}

	validateReq := models.VerifyTokenRequest{
		Token: tokens.AccessToken,
	}

	body, _ := json.Marshal(validateReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/validate-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateToken(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use dedicated handler and service for this test
	secretKey := getTestSecretKey()
	accessTokenDuration := int64(15 * 60)
	refreshTokenDuration := int64(7 * 24 * 3600)

	jwtService := services.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	handler := handlers.NewAuthHandler(jwtService)

	loginReq := models.LoginRequest{
		ID:             "user789",
		Email:          "user@example.com",
		EMAIL_VERIFIED: true,
		Roles:          []string{"user"},
	}

	tokens, err := jwtService.Login(&loginReq)
	if err != nil {
		t.Fatalf("failed to get tokens: %v", err)
	}

	refreshReq := models.RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	}

	body, _ := json.Marshal(refreshReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
