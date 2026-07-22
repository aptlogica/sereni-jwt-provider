/*
Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
This file is part of software developed by Aptlogica Technologies Private Limited.
Licensed under the Apache License, Version 2.0. See the LICENSE file in the project root
for full license information.
Websites:
https://www.aptlogica.com
https://www.serenibase.com
Support:
support@aptlogica.com
support@serenibase.com
*/

package main

import (
	"log"
	"os"
	"strings"

	"github.com/aptlogica/sereni-jwt-provider/internal/config"
	"github.com/aptlogica/sereni-jwt-provider/internal/handlers"
	"github.com/aptlogica/sereni-jwt-provider/internal/services"

	_ "github.com/aptlogica/sereni-jwt-provider/docs/swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           JWT Authentication Service API
// @version         1.0
// @description     A Keycloak-like JWT-based authentication service built with Go and Gin
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8081
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Validate JWT secret strength at startup
	if err := config.ValidateJWTSecret(cfg.JWTSecret); err != nil {
		log.Fatalf("JWT_SECRET validation failed: %v", err)
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}

	// Initialize dependencies
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	authHandler := handlers.NewAuthHandler(jwtService)

	// Setup Gin router
	router := gin.Default()

	// CORS Setup
	router.Use(cors.New(cors.Config{
		// AllowAllOrigins:  true,
		AllowOrigins:     strings.Split(cfg.AllowedOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check endpoint (only root /health kept)
	health := router.Group("/health")
	{
		health.GET("", authHandler.Health)
	}

	// Auth endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/login", authHandler.Login)
		auth.POST("/validate-token", authHandler.ValidateToken)
		auth.POST("/verify-token", authHandler.VerifyToken)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	serverAddr := cfg.ServerHost + ":" + cfg.ServerPort
	log.Printf("Starting auth service on %s", serverAddr)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/index.html", cfg.ServerPort)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
