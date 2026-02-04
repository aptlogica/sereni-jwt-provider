package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"log"
	"os"

	_ "auth-service/docs/swagger"

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
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required — refusing to start")
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}

	// Initialize dependencies
	tokenStore := repository.NewTokenStore()
	tokenStore.StartCleanupRoutine()
	defer tokenStore.Close()
	jwtService := services.NewJWTService(cfg.JWTSecret, tokenStore)
	authHandler := handlers.NewAuthHandler(jwtService)

	// Setup Gin router
	router := gin.Default()

	// CORS Setup
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
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
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(jwtService), authHandler.Logout)
		auth.POST("/validate-token", authHandler.ValidateToken)
		auth.POST("/verify-token", authHandler.VerifyToken)
	}

	// Protected example endpoint
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		protected.GET("/profile", authHandler.GetProfile)
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
