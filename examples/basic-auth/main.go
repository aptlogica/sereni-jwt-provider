package main

import (
	"fmt"
	"log"
	"time"

	jwt "github.com/aptlogica/sereni-jwt-provider"
)

func main() {
	fmt.Println("=== Sereni JWT Provider - Basic Authentication Example ===")

	// Initialize JWT provider with configuration
	config := jwt.Config{
		Secret:    "my-secret-key-for-development-only",
		Expiry:    time.Hour * 24, // 24 hours
		Issuer:    "sereni-jwt-provider",
		Algorithm: "HS256",
	}

	provider := jwt.NewProvider(config)

	// Step 1: Generate a JWT token for a user
	fmt.Println("\n1. Generating JWT token...")

	userClaims := map[string]interface{}{
		"user_id":  "12345",
		"username": "john_doe",
		"email":    "john@example.com",
		"role":     "user",
	}

	token, err := provider.GenerateToken(userClaims)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Printf("✅ Token generated successfully!\n")
	fmt.Printf("Token: %s...\n", token[:50])

	// Step 2: Validate the token
	fmt.Println("\n2. Validating JWT token...")

	claims, err := provider.ValidateToken(token)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}

	fmt.Printf("✅ Token is valid!\n")
	fmt.Printf("User ID: %v\n", claims["user_id"])
	fmt.Printf("Username: %v\n", claims["username"])
	fmt.Printf("Email: %v\n", claims["email"])
	fmt.Printf("Role: %v\n", claims["role"])

	// Step 3: Test with expired token (demonstration)
	fmt.Println("\n3. Testing token expiration...")

	// Create a provider with very short expiry for demo
	shortConfig := jwt.Config{
		Secret:    "my-secret-key-for-development-only",
		Expiry:    time.Second * 1, // 1 second
		Issuer:    "sereni-jwt-provider",
		Algorithm: "HS256",
	}

	shortProvider := jwt.NewProvider(shortConfig)
	shortToken, _ := shortProvider.GenerateToken(userClaims)

	fmt.Println("Waiting for token to expire...")
	time.Sleep(time.Second * 2)

	_, err = shortProvider.ValidateToken(shortToken)
	if err != nil {
		fmt.Printf("✅ Expected error - token expired: %v\n", err)
	}

	// Step 4: Test with invalid token
	fmt.Println("\n4. Testing invalid token...")

	invalidToken := "invalid.jwt.token"
	_, err = provider.ValidateToken(invalidToken)
	if err != nil {
		fmt.Printf("✅ Expected error - invalid token: %v\n", err)
	}

	fmt.Println("\n=== Example completed successfully! ===")
	fmt.Println("\nNext steps:")
	fmt.Println("- Check out middleware example for HTTP route protection")
	fmt.Println("- See refresh-tokens example for token rotation")
	fmt.Println("- Explore multi-tenant example for advanced use cases")
}
