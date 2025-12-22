JWT Authentication Service
A production-grade, Keycloak-like JWT-based authentication service built with Go and Gin framework.

Features
✅ JWT-based authentication (Access + Refresh tokens)
✅ Token rotation on refresh
✅ User registration and login
✅ Token validation and verification
✅ Protected endpoints with middleware
✅ Role-based access control
✅ Complete Swagger/OpenAPI documentation
✅ CORS support with configurable origins
✅ Health check endpoints
✅ In-memory token store (thread-safe)
✅ Bcrypt password hashing
✅ Clean architecture
Tech Stack
Go 1.21+
Gin - HTTP web framework
JWT (golang-jwt/jwt/v5) - JSON Web Tokens
Swagger (swaggo) - API documentation
Bcrypt - Password hashing
Project Structure
auth-service/
├── main.go                          # Entry point
├── go.mod                           # Go modules
├── Makefile                         # Build commands
├── Dockerfile                       # Docker configuration
├── README.md                        # This file
├── docs/                            # Generated Swagger docs
└── internal/
    ├── handlers/
    │   └── auth_handler.go         # HTTP handlers
    ├── services/
    │   └── jwt_service.go          # Business logic
    ├── middleware/
    │   └── auth_middleware.go      # Auth middleware
    ├── models/
    │   └── auth_models.go          # Data models
    ├── repository/
    │   └── token_store.go          # Token storage
    └── utils/
        ├── response.go             # Response helpers
        └── password.go             # Password utilities
Quick Start
Prerequisites
Go 1.21 or higher
Make (optional)
Installation
Clone the repository
bash
git clone <repository-url>
cd auth-service
Install dependencies
bash
go mod download
go mod tidy
Install Swagger CLI
bash
go install github.com/swaggo/swag/cmd/swag@latest
Generate Swagger documentation
bash
swag init -g main.go --output ./docs
Run the service
bash
export JWT_SECRET="your-super-secret-key-change-in-production"
export PORT=8080
export ALLOWED_ORIGINS="http://localhost:3000,http://localhost:5173"
go run main.go
Or using Makefile:

bash
make install
make run
Using Docker
bash
# Build image
docker build -t auth-service:latest .

# Run container
docker run -p 8080:8080 \
  -e JWT_SECRET=my-super-secret-key \
  -e ALLOWED_ORIGINS="http://localhost:3000,https://myapp.com" \
  auth-service:latest
API Endpoints
Health Endpoints
GET /health - Service health check
GET /health/live - Liveness probe
GET /health/ready - Readiness probe
Authentication Endpoints
POST /auth/register - Register new user
POST /auth/login - Login and get tokens
POST /auth/refresh - Refresh access token
POST /auth/logout - Logout (revoke refresh token)
POST /auth/validate-token - Validate token and get claims
POST /auth/verify-token - Verify token validity (boolean)
Protected Endpoints
GET /api/profile - Get user profile (requires authentication)
Swagger UI
Access the interactive API documentation at:

http://localhost:8080/swagger/index.html
API Usage Examples
1. Register a new user
bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "roles": ["user"]
  }'
Response:

json
{
  "success": true,
  "code": "USER_CREATED",
  "message": "User registered successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "roles": ["user"]
  }
}
2. Login
bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
Response:

json
{
  "success": true,
  "code": "LOGIN_SUCCESS",
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
3. Access Protected Endpoint
bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
Response:

json
{
  "success": true,
  "code": "PROFILE_RETRIEVED",
  "message": "Profile retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "roles": ["user"]
  }
}
4. Refresh Token
bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
Response:

json
{
  "success": true,
  "code": "TOKEN_REFRESHED",
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "NEW_ACCESS_TOKEN",
    "refresh_token": "NEW_REFRESH_TOKEN",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
5. Validate Token
bash
curl -X POST http://localhost:8080/auth/validate-token \
  -H "Content-Type: application/json" \
  -d '{
    "token": "YOUR_TOKEN"
  }'
Response:

json
{
  "success": true,
  "code": "TOKEN_VALID",
  "message": "Token is valid",
  "data": {
    "sub": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "roles": ["user"],
    "token_type": "access",
    "exp": 1703001600,
    "iat": 1703000700,
    "iss": "auth-service"
  }
}
6. Logout
bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
Response:

json
{
  "success": true,
  "code": "LOGOUT_SUCCESS",
  "message": "Logged out successfully",
  "data": null
}
JWT Token Design
Token Types
Access Token
Duration: 15 minutes
Used for API authentication
Short-lived for security
Refresh Token
Duration: 7 days
Used to obtain new access tokens
Stored in token store
Rotated on refresh
JWT Claims
json
{
  "sub": "user_id",
  "email": "user@example.com",
  "roles": ["user", "admin"],
  "token_type": "access",
  "iss": "auth-service",
  "iat": 1703000700,
  "exp": 1703001600,
  "nbf": 1703000700
}
Security Features
✅ HS256 signing algorithm
✅ Configurable JWT secret
✅ Token expiration validation
✅ Token type validation (access vs refresh)
✅ Refresh token rotation
✅ Refresh token blacklisting on logout
✅ Thread-safe token storage
Configuration
Configure the service using environment variables:

bash
# Required
export JWT_SECRET="your-secret-key-min-32-chars"

# Optional
export PORT="8080"                    # Default: 8080
export ALLOWED_ORIGINS="http://localhost:3000,https://myapp.com"  # Default: localhost:3000,8080,5173
CORS Configuration
The service includes CORS support with configurable origins. You can set allowed origins in two ways:

Environment Variable (recommended for production):
bash
export ALLOWED_ORIGINS="http://localhost:3000,https://myapp.com,https://admin.myapp.com"
Default (for development): If ALLOWED_ORIGINS is not set, the service allows these origins by default:
http://localhost:3000 (React, Next.js)
http://localhost:8080 (Same origin)
http://localhost:5173 (Vite)
CORS Configuration includes:

✅ Configurable allowed origins
✅ Common HTTP methods (GET, POST, PUT, PATCH, DELETE, OPTIONS)
✅ Authorization header support
✅ Credentials support
✅ 12-hour preflight cache
Development
Run Tests
bash
go test -v ./...
Generate Swagger Docs
bash
swag init -g main.go --output ./docs
Build Binary
bash
go build -o bin/auth-service main.go
Using Makefile
bash
make help      # Show available commands
make install   # Install dependencies
make swagger   # Generate Swagger docs
make run       # Run the service
make build     # Build binary
make clean     # Clean build artifacts
make test      # Run tests
Production Considerations
Security
Set a strong JWT_SECRET (minimum 32 characters)
Use HTTPS in production
Configure CORS properly - set specific allowed origins, not wildcards
Implement rate limiting for auth endpoints
Enable request logging
Implement token refresh rotation (already included)
Add database persistence for production use
Scalability
Replace in-memory store with Redis/PostgreSQL
Implement distributed caching
Add database connection pooling
Use environment-specific configurations
Add metrics and monitoring (Prometheus)
Implement graceful shutdown
Monitoring
Add health check dependencies verification
Implement structured logging (logrus/zap)
Add request ID tracing
Monitor token generation/validation metrics
Set up alerting for failures
Response Format
All API responses follow this format:

json
{
  "success": true,
  "code": "RESPONSE_CODE",
  "message": "Human readable message",
  "data": {
    // Response data
  }
}
HTTP Status Codes
200 OK - Successful request
201 Created - Resource created
400 Bad Request - Invalid request
401 Unauthorized - Authentication required
403 Forbidden - Insufficient permissions
404 Not Found - Resource not found
409 Conflict - Resource already exists
500 Internal Server Error - Server error
License
Apache 2.0

Support
For issues and questions, please open an issue on the repository.

