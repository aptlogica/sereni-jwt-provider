// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package handlers

import (
	"fmt"
	serrors "github.com/aptlogica/sereni-jwt-provider/internal/errors"
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"github.com/aptlogica/sereni-jwt-provider/internal/services"
	"github.com/aptlogica/sereni-jwt-provider/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	jwtService *services.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and receive JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Login credentials"
// @Success      200 {object} models.SuccessResponse{data=models.TokenResponse}
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.jwtService.Login(&req)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	utils.SendSuccess(c, "LOGIN_SUCCESS", "Login successful", tokens)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Generate a new access token using a refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} models.SuccessResponse{data=models.TokenResponse}
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		fmt.Println("err: ", err)
		utils.SendError(c, http.StatusUnauthorized, "REFRESH_FAILED", "Invalid or expired refresh token")
		return
	}

	utils.SendSuccess(c, "TOKEN_REFRESHED", "Token refreshed successfully", tokens)
}

// ValidateToken godoc
// @Summary      Validate token
// @Description  Validate a JWT token and return its claims
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.VerifyTokenRequest true "Token to validate"
// @Success      200 {object} models.SuccessResponse{data=models.TokenClaims}
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /auth/validate-token [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	claims, err := h.jwtService.ValidateToken(req.Token, true)
	fmt.Printf("[DEBUG] ValidateToken: token=%s, claims=%+v, err=%v\n", req.Token, claims, err)
	if err != nil || claims == nil {
		if err == serrors.ErrTokenExpire {
			utils.SendError(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired")
			return
		}
		if err == serrors.ErrInvalidToken {
			utils.SendError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid")
			return
		}
		utils.SendError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
		return
	}

	// Convert to Swagger-friendly format
	tokenClaims := models.TokenClaims{}
	if claims != nil {
		tokenClaims.UserID = claims.UserID
		tokenClaims.Email = claims.Email
		tokenClaims.Roles = claims.Roles
		tokenClaims.TokenType = claims.TokenType
		tokenClaims.Issuer = claims.Issuer
		tokenClaims.Subject = claims.Subject
		if claims.ExpiresAt != nil {
			tokenClaims.ExpiresAt = claims.ExpiresAt.Unix()
		}
		if claims.IssuedAt != nil {
			tokenClaims.IssuedAt = claims.IssuedAt.Unix()
		}
		if claims.NotBefore != nil {
			tokenClaims.NotBefore = claims.NotBefore.Unix()
		}
	}

	utils.SendSuccess(c, "TOKEN_VALID", "Token is valid", tokenClaims)
}

// VerifyToken godoc
// @Summary      Verify token
// @Description  Verify if a token is valid (returns boolean)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.VerifyTokenRequest true "Token to verify"
// @Success      200 {object} models.SuccessResponse{data=map[string]bool}
// @Failure      400 {object} models.ErrorResponse
// @Router       /auth/verify-token [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req models.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	_, err := h.jwtService.ValidateToken(req.Token, true)
	isValid := err == nil

	result := map[string]bool{
		"valid": isValid,
	}

	utils.SendSuccess(c, "TOKEN_VERIFIED", "Token verification complete", result)
}

// Health godoc
// @Summary      Health check
// @Description  Check if the service is running
// @Tags         health
// @Produce      json
// @Success      200 {object} models.SuccessResponse{data=map[string]string}
// @Router       /health [get]
func (h *AuthHandler) Health(c *gin.Context) {
	utils.SendSuccess(c, "HEALTHY", "Service is healthy", map[string]string{
		"status": "healthy",
	})
}
