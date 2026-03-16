package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/aptlogica/sereni-jwt-provider"
)

var provider *jwt.Provider

func main() {
	fmt.Println("=== Sereni JWT Provider - Middleware Example ===")

	// Initialize JWT provider
	config := jwt.Config{
		Secret:    "my-secret-key-for-development-only",
		Expiry:    time.Hour * 24,
		Issuer:    "sereni-jwt-provider",
		Algorithm: "HS256",
	}

	provider = jwt.NewProvider(config)

	// Setup routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/protected", jwtMiddleware(protectedHandler))
	http.HandleFunc("/public", publicHandler)
	http.HandleFunc("/profile", jwtMiddleware(profileHandler))

	fmt.Println("\n🚀 Server starting on :8080")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  POST /login      - Get JWT token (email: user@example.com, password: password123)")
	fmt.Println("  GET  /public     - Public endpoint (no auth required)")
	fmt.Println("  GET  /protected  - Protected endpoint (requires JWT)")
	fmt.Println("  GET  /profile    - Get user profile (requires JWT)")
	fmt.Println("\nExample usage:")
	fmt.Println(`  curl -X POST -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"password123"}' http://localhost:8080/login`)
	fmt.Println(`  curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/protected`)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Login handler - generates JWT token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simple authentication (in real app, check database)
	if loginReq.Email != "user@example.com" || loginReq.Password != "password123" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	claims := map[string]interface{}{
		"user_id":  "12345",
		"username": "john_doe",
		"email":    loginReq.Email,
		"role":     "user",
	}

	token, err := provider.GenerateToken(claims)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token":      token,
		"expires_in": 86400, // 24 hours in seconds
		"token_type": "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// JWT middleware - protects routes
func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := provider.ValidateToken(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add claims to request context (in real app, use context.Context)
		r.Header.Set("X-User-ID", fmt.Sprintf("%v", claims["user_id"]))
		r.Header.Set("X-Username", fmt.Sprintf("%v", claims["username"]))
		r.Header.Set("X-User-Email", fmt.Sprintf("%v", claims["email"]))
		r.Header.Set("X-User-Role", fmt.Sprintf("%v", claims["role"]))

		// Call next handler
		next(w, r)
	}
}

// Public endpoint - no authentication required
func publicHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "This is a public endpoint",
		"timestamp": time.Now().Unix(),
		"auth":      "not required",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Protected endpoint - requires valid JWT
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "This is a protected endpoint",
		"timestamp": time.Now().Unix(),
		"auth":      "required",
		"user_id":   r.Header.Get("X-User-ID"),
		"username":  r.Header.Get("X-Username"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Profile endpoint - returns user information from JWT
func profileHandler(w http.ResponseWriter, r *http.Request) {
	profile := map[string]interface{}{
		"user_id":  r.Header.Get("X-User-ID"),
		"username": r.Header.Get("X-Username"),
		"email":    r.Header.Get("X-User-Email"),
		"role":     r.Header.Get("X-User-Role"),
		"profile":  "active",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
