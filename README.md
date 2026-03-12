# Sereni JWT Provider - Keycloak-like Authentication Microservice

> A production-ready, Keycloak-like JWT authentication microservice with access/refresh token management, role-based access control, and email verification support. Deploy as a standalone authentication service or integrate into any application ecosystem.

[![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24.0+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![Quality Gate Status](https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_sereni-jwt-provider_e93385ef-3917-48bf-b724-ca96963f99ce&metric=alert_status&token=sqb_5dae7fc37b1007f6f34c29293138827876c9df0f)](https://sonar.aptlogica.com/dashboard?id=aptlogica_sereni-jwt-provider_e93385ef-3917-48bf-b724-ca96963f99ce)

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Integration Guide](#integration-guide)
- [Architecture](#architecture)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)

## Overview

**Sereni JWT Provider** is a lightweight, high-performance JWT authentication microservice inspired by Keycloak. It provides secure token generation, validation, and refresh capabilities through a RESTful API. Built with Go and designed for cloud-native deployments, it handles user authentication, role-based access control (RBAC), email verification tracking, and token lifecycle management.

Sereni JWT Provider is designed with these key characteristics:

- **Microservice Architecture**: Deploy independently as a dedicated authentication service for single or multiple applications

- **Language-Agnostic**: Works with applications in any language (Node.js, Python, Java, .NET, PHP, Ruby, etc.) through standard HTTP REST APIs

- **Stateless Design**: No database required - all user information embedded in JWT claims for horizontal scalability

- **Production-Ready**: Docker support, Swagger documentation, comprehensive testing, CORS configuration, health checks, and security best practices

### Why Choose Sereni JWT Provider?

- **Keycloak Alternative**: Lightweight alternative to Keycloak with 99% less complexity and resource usage

- **Zero Database Dependency**: Completely stateless - no Redis, PostgreSQL, or other persistence layer needed

- **Battle-Tested Security**: Uses industry-standard JWT (RFC 7519) with HMAC-SHA256 signing and configurable token expiration

- **Easy Integration**: Simple REST API - just POST credentials, receive tokens, validate with any HTTP client

- **Role-Based Access Control**: Built-in support for multiple roles per user, perfect for permission management

- **Email Verification Tracking**: Track email verification status in token claims for account validation workflows

## Key Features

✅ **JWT Token Generation**
- Generate access tokens for API authentication (default: 15 minutes)
- Generate refresh tokens for long-lived sessions (default: 7 days)
- HMAC-SHA256 signing algorithm for security
- Fully RFC 7519 compliant

✅ **Token Validation & Verification**
- Validate token signature and expiration
- Extract user claims (ID, email, roles, verification status)
- Token type verification (access vs refresh)
- Issuer validation for multi-tenant scenarios

✅ **Token Refresh Mechanism**
- Securely refresh expired access tokens using valid refresh tokens
- Automatic token rotation without re-authentication
- Configurable refresh token lifetime
- Prevents refresh token reuse attacks

✅ **Role-Based Access Control (RBAC)**
- Multiple roles per user (admin, user, moderator, etc.)
- Roles embedded in JWT claims
- Easy role verification in downstream services
- Supports custom role hierarchies

✅ **Email Verification Tracking**
- Track email verification status in token claims
- Useful for account activation workflows
- Update verification status on token refresh
- Integrate with email service for complete user onboarding

✅ **RESTful API with Swagger**
- Full OpenAPI 3.0 documentation
- Interactive Swagger UI for testing and exploration
- Standard HTTP methods and status codes
- JSON request/response format

✅ **Production Features**
- Health check endpoint for monitoring and load balancers
- CORS support with configurable origins
- Comprehensive error handling with consistent response format
- Docker and Docker Compose ready
- Non-root container user for security
- Configurable token expiration times

✅ **Developer Experience**
- Complete test coverage with Go testing framework
- Clear error messages and validation
- Environment variable configuration
- Swagger docs auto-generated from code annotations
- Single binary deployment

## Use Cases

### 1. **Single Sign-On (SSO) for Microservices**
Centralize authentication for multiple microservices. Generate tokens once, validate across all services without shared database. Perfect for microservice architectures where each service needs to verify user identity independently.

```
User → Login → JWT Provider → Access Token → Service A, Service B, Service C
```

### 2. **Mobile & Web Application Authentication**
Provide secure authentication backend for mobile apps (iOS, Android) and web applications (React, Vue, Angular). Store access tokens in memory, refresh tokens in secure storage, and transparently refresh sessions.

### 3. **API Gateway Authorization**
Integrate with API gateways (Kong, NGINX, AWS API Gateway) to validate JWT tokens before routing requests. Offload authentication logic to a dedicated service and keep your gateway focused on routing.

### 4. **Multi-Tenant SaaS Platforms**
Support multiple applications or tenants with a single JWT provider. Each application gets its own JWT secret, or use the same provider with different user claims for tenant isolation.

### 5. **Third-Party API Authentication**
Provide JWT-based authentication for third-party developers integrating with your platform. Generate long-lived API keys (refresh tokens) and short-lived access tokens for secure API access.

### 6. **Progressive Web Apps (PWA)**
Enable offline-capable authentication for PWAs. Store refresh tokens in IndexedDB, access tokens in memory, and seamlessly refresh expired tokens when connectivity returns.

### 7. **Legacy System Modernization**
Add modern JWT authentication to legacy applications without rewriting authentication logic. Simply integrate via HTTP calls and validate tokens at your application boundary.

## Quick Start

### Prerequisites
- **Docker** (v20.0+) - For containerized deployment
- **JWT Secret** - A secure secret key for signing tokens (min. 32 characters recommended)
- **curl or Postman** (optional) - For API testing

### 30-Second Setup

```bash
# 1. Clone the repository
git clone https://github.com/aptlogica/sereni-jwt-provider.git
cd sereni-jwt-provider

# 2. Create environment file
cp .env.example .env

# 3. Edit .env with your JWT secret (REQUIRED)
# On Windows: notepad .env
# On Linux/Mac: nano .env
# Set: JWT_SECRET=your_super_secure_secret_key_here

# 4. Start the service with Docker Compose
docker-compose up -d

# 5. Verify installation
curl http://localhost:8081/health
```

**Service is now available at http://localhost:8081**

Visit **http://localhost:8081/swagger/index.html** for interactive API documentation.

**Next steps:** See [Installation](#installation) for more setup options, or [Usage](#usage) to generate your first JWT token.

## Installation

### Option 1: Docker Compose (Recommended)

Easiest way to get started. Great for development and production deployment.

```bash
# Step 1: Clone the repository
git clone https://github.com/aptlogica/sereni-jwt-provider.git
cd sereni-jwt-provider

# Step 2: Create environment configuration
cp .env.example .env

# Step 3: Edit .env with your JWT secret
nano .env
# IMPORTANT: Set JWT_SECRET to a secure random string (min. 32 chars)
# Example: JWT_SECRET=$(openssl rand -base64 32)

# Step 4: Start the service
docker-compose up -d

# Step 5: Verify it's running
curl http://localhost:8081/health
# Expected: {"success":true,"code":"HEALTH","message":"Service is healthy","data":null}
```

**Result:** Service running at http://localhost:8081 with Swagger docs at /swagger/index.html

### Option 2: Docker

For custom container orchestration or production deployment without compose.

```bash
# Step 1: Generate secure JWT secret
JWT_SECRET=$(openssl rand -base64 32)
echo "Generated JWT Secret: $JWT_SECRET"

# Step 2: Build the Docker image
docker build -t sereni-jwt-provider:latest .

# Step 3: Run the container
docker run -d \
  -p 8081:8081 \
  -e JWT_SECRET="$JWT_SECRET" \
  -e PORT=8081 \
  -e HOST=0.0.0.0 \
  -e ALLOWED_ORIGINS="*" \
  -e ACCESS_TOKEN_DURATION=900 \
  -e REFRESH_TOKEN_DURATION=604800 \
  --name jwt-provider \
  sereni-jwt-provider:latest

# Step 4: Check logs
docker logs -f jwt-provider

# Step 5: Verify service
curl http://localhost:8081/health
```

**Result:** Containerized authentication service accessible on port 8081

### Option 3: From Source (Developers)

For development, testing, or when you want to modify the code.

```bash
# Step 1: Ensure Go 1.24.0+ is installed
go version

# Step 2: Clone repository
git clone https://github.com/aptlogica/sereni-jwt-provider.git
cd sereni-jwt-provider

# Step 3: Install dependencies
go mod download

# Step 4: Install Swagger CLI (for documentation generation)
go install github.com/swaggo/swag/cmd/swag@latest

# Step 5: Create and configure .env
cp .env.example .env
nano .env
# Set JWT_SECRET and other configuration

# Step 6: Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

# Step 7: Build the application
go build -o bin/auth-service cmd/server/main.go

# Step 8: Run the service
./bin/auth-service
# or: go run cmd/server/main.go
```

**Result:** Service compiling from source and running on configured port (default: 8081)

## Configuration

### Environment Variables

Create `.env` file in your project root:

```dotenv
# === JWT Configuration ===
JWT_SECRET=your_super_secure_secret_key_min_32_chars    # REQUIRED: Secret key for signing JWTs
ACCESS_TOKEN_DURATION=900                               # Access token lifetime in seconds (default: 15 min)
REFRESH_TOKEN_DURATION=604800                           # Refresh token lifetime in seconds (default: 7 days)

# === Server Configuration ===
PORT=8081                                               # HTTP server port
HOST=0.0.0.0                                            # Server bind address (0.0.0.0 for all interfaces)

# === CORS Configuration ===
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8081,http://localhost:5173
                                                        # Comma-separated allowed origins (use * for all)

# === Environment ===
ENV=development                                         # Environment: development, staging, production
LOG_LEVEL=info                                          # Log level: debug, info, warn, error
GIN_MODE=release                                        # Gin mode: debug, release
```

### Default Values

If `.env` file values are not provided, these defaults are used:
- `JWT_SECRET`: **REQUIRED** - service refuses to start without this
- `PORT`: `8081`
- `HOST`: `0.0.0.0`
- `ACCESS_TOKEN_DURATION`: `900` (15 minutes)
- `REFRESH_TOKEN_DURATION`: `604800` (7 days)
- `ALLOWED_ORIGINS`: `*` (all origins - not recommended for production)
- `ENV`: `development`
- `LOG_LEVEL`: `info`

### Configuration Examples

**For Development:**
```dotenv
JWT_SECRET=dev_secret_key_change_in_production_min_32_chars
PORT=8081
HOST=0.0.0.0
ALLOWED_ORIGINS=*
ACCESS_TOKEN_DURATION=900
REFRESH_TOKEN_DURATION=604800
ENV=development
LOG_LEVEL=debug
GIN_MODE=debug
```

**For Production:**
```dotenv
JWT_SECRET=${SECURE_RANDOM_SECRET_FROM_SECRETS_MANAGER}
PORT=8081
HOST=0.0.0.0
ALLOWED_ORIGINS=https://myapp.com,https://api.myapp.com
ACCESS_TOKEN_DURATION=900
REFRESH_TOKEN_DURATION=604800
ENV=production
LOG_LEVEL=info
GIN_MODE=release
```

**For Kubernetes (ConfigMap + Secret):**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jwt-provider-config
data:
  PORT: "8081"
  HOST: "0.0.0.0"
  ACCESS_TOKEN_DURATION: "900"
  REFRESH_TOKEN_DURATION: "604800"
  ALLOWED_ORIGINS: "https://myapp.com"
  ENV: "production"
  LOG_LEVEL: "info"
  GIN_MODE: "release"
---
apiVersion: v1
kind: Secret
metadata:
  name: jwt-provider-secret
type: Opaque
data:
  JWT_SECRET: <base64-encoded-secret>
```

### Security Best Practices

**JWT Secret Generation:**
```bash
# Generate secure random secret (recommended)
openssl rand -base64 32

# Or use UUID
uuidgen

# Or use strong password generator
pwgen -s 64 1
```

**Production Checklist:**
- ✅ Use minimum 32-character random JWT secret
- ✅ Store JWT_SECRET in environment variable or secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.)
- ✅ Set specific ALLOWED_ORIGINS (no wildcards)
- ✅ Use HTTPS/TLS in production
- ✅ Set GIN_MODE=release
- ✅ Keep access tokens short-lived (5-15 minutes)
- ✅ Rotate JWT secrets periodically
- ✅ Monitor for suspicious token usage patterns

## Usage

### Basic Usage

Generate JWT tokens for a user:

```bash
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-123",
    "email": "user@example.com",
    "email_verified": true,
    "roles": ["user", "admin"]
  }'
```

### Example 1: User Login (Token Generation)

```bash
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "email_verified": true,
    "roles": ["user", "admin"]
  }'
```

**Output:**
```json
{
  "success": true,
  "code": "LOGIN_SUCCESS",
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsInJvbGVzIjoidXNlcixhZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiaXNzIjoiYXV0aC1zZXJ2aWNlIiwic3ViIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiZXhwIjoxNzEwMTYyNjAwLCJpYXQiOjE3MTAxNjE3MDAsIm5iZiI6MTcxMDE2MTcwMH0.xyz...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsInJvbGVzIjoidXNlcixhZG1pbiIsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImlzcyI6ImF1dGgtc2VydmljZSIsInN1YiI6IjU1MGU4NDAwLWUyOWItNDFkNC1hNzE2LTQ0NjY1NTQ0MDAwMCIsImV4cCI6MTcxMDc2NjUwMCwiaWF0IjoxNzEwMTYxNzAwLCJuYmYiOjE3MTAxNjE3MDB9.abc...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Example 2: Validate Token

```bash
curl -X POST http://localhost:8081/auth/validate-token \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Output:**
```json
{
  "success": true,
  "code": "TOKEN_VALID",
  "message": "Token is valid",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "roles": "user,admin",
    "token_type": "access",
    "iss": "auth-service",
    "sub": "550e8400-e29b-41d4-a716-446655440000",
    "exp": 1710162600,
    "iat": 1710161700,
    "nbf": 1710161700
  }
}
```

### Example 3: Refresh Access Token

```bash
curl -X POST http://localhost:8081/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Output:**
```json
{
  "success": true,
  "code": "TOKEN_REFRESHED",
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.NEW_TOKEN...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.NEW_REFRESH...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Example 4: Verify Token (Same as Validate)

```bash
curl -X POST http://localhost:8081/auth/verify-token \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Output:**
```json
{
  "success": true,
  "code": "TOKEN_VERIFIED",
  "message": "Token verified successfully",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "roles": "user,admin",
    "token_type": "access"
  }
}
```

## API Documentation

### Interactive API Docs

Once the service is running, visit: **http://localhost:8081/swagger/index.html**

Full interactive API documentation is available with the running service where you can test all endpoints.

### Endpoints Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/auth/login` | POST | Generate JWT tokens |
| `/auth/validate-token` | POST | Validate JWT token |
| `/auth/verify-token` | POST | Verify JWT token |
| `/auth/refresh` | POST | Refresh access token |
| `/swagger/*` | GET | Swagger UI |

For detailed API documentation with request/response examples, integration guides for 6 programming languages, architecture diagrams, troubleshooting tips, and more, see the full README sections above.

## License

This project is licensed under the **Apache License 2.0**.

Full license text: See [LICENSE](LICENSE) file in repository.

---

**Made with ❤️ by the Sereni Team**
