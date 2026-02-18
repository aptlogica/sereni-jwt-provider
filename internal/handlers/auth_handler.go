package handlers

import (
	"auth-service/internal/errors"
	"auth-service/internal/models"
	"auth-service/internal/services"
	"auth-service/internal/utils"
	"fmt"
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

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "Registration details"
// @Success      201 {object} models.SuccessResponse{data=models.UserProfile}
// @Failure      400 {object} models.ErrorResponse
// @Failure      409 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.jwtService.Register(req.UserID, req.Email, req.Password, req.Roles)
	if err != nil {

		if err == errors.ErrUserExists {
			utils.SendError(c, http.StatusConflict, "USER_EXISTS", "User with this email already exists")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "Failed to register user")
		return
	}

	profile := models.UserProfile{
		ID:    user.ID,
		Email: user.Email,
		Roles: user.Roles,
	}

	utils.SendCreated(c, "USER_CREATED", "User registered successfully", profile)
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

	tokens, err := h.jwtService.Login(req.Email, req.Password)
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
		utils.SendError(c, http.StatusUnauthorized, "REFRESH_FAILED", "Invalid or expired refresh token")
		return
	}

	utils.SendSuccess(c, "TOKEN_REFRESHED", "Token refreshed successfully", tokens)
}

// Logout godoc
// @Summary      Logout user
// @Description  Revoke refresh token and logout user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.RefreshTokenRequest true "Refresh token to revoke"
// @Success      200 {object} models.SuccessResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.jwtService.Logout(req.RefreshToken); err != nil {
		utils.SendError(c, http.StatusBadRequest, "LOGOUT_FAILED", "Failed to revoke token")
		return
	}

	utils.SendSuccess(c, "LOGOUT_SUCCESS", "Logged out successfully", nil)
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

	claims, err := h.jwtService.ValidateToken(req.Token)
	fmt.Println("err: ", err)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
		return
	}

	// Convert to Swagger-friendly format
	tokenClaims := models.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Roles:     claims.Roles,
		TokenType: claims.TokenType,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		ExpiresAt: claims.ExpiresAt.Unix(),
		IssuedAt:  claims.IssuedAt.Unix(),
		NotBefore: claims.NotBefore.Unix(),
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

	_, err := h.jwtService.ValidateToken(req.Token)
	isValid := err == nil

	result := map[string]bool{
		"valid": isValid,
	}

	utils.SendSuccess(c, "TOKEN_VERIFIED", "Token verification complete", result)
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get authenticated user's profile information
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.SuccessResponse{data=models.UserProfile}
// @Failure      401 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /api/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required")
		return
	}

	profile, err := h.jwtService.GetUserProfile(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	utils.SendSuccess(c, "PROFILE_RETRIEVED", "Profile retrieved successfully", profile)
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

// HealthLive godoc
// (Removed HealthLive and HealthReady handlers; only Health remains)
