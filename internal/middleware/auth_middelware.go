package middleware

import (
	"auth-service/internal/services"
	"auth-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "AUTH_MISSING", "Authorization header is required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.SendError(c, http.StatusUnauthorized, "AUTH_INVALID_FORMAT", "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Invalid or expired token")
			c.Abort()
			return
		}

		// Check if it's an access token
		if claims.TokenType != services.TokenTypeAccess {
			utils.SendError(c, http.StatusUnauthorized, "AUTH_WRONG_TOKEN_TYPE", "Access token required")
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			utils.SendError(c, http.StatusForbidden, "AUTH_NO_ROLES", "No roles found")
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			utils.SendError(c, http.StatusForbidden, "AUTH_INVALID_ROLES", "Invalid roles format")
			c.Abort()
			return
		}

		hasRole := false
		for _, r := range userRoles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.SendError(c, http.StatusForbidden, "AUTH_INSUFFICIENT_PERMISSIONS", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
