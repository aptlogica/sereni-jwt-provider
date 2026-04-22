<h1 align="center">sereni-jwt-provider - Secure JWT Authentication Service</h1>

<p align="center">Enterprise-grade JWT authentication service and open source auth provider for secure application access. A comprehensive JWT auth server and identity provider offering advanced token management, key rotation, and seamless integration with modern authentication workflows.</p>

<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version"></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26.2-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
<a href="https://jwt.io/"><img src="https://img.shields.io/badge/JWT-Authentication-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white" alt="JWT"></a>
<a href="https://gin-gonic.com/"><img src="https://img.shields.io/badge/Gin-Framework-008ECF?style=for-the-badge&logo=gin&logoColor=white" alt="Gin"></a>
<a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></a>
<a href="https://swagger.io/"><img src="https://img.shields.io/badge/Swagger-Documented-85EA2D?style=for-the-badge&logo=swagger&logoColor=black" alt="Swagger"></a>
</p>

<p align="center">
<a href="https://github.com/aptlogica/sereni-jwt-provider/actions/workflows/ci.yml"><img src="https://github.com/aptlogica/sereni-jwt-provider/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
<a href="https://github.com/aptlogica/sereni-jwt-provider/actions/workflows/github-code-scanning/codeql"><img src="https://github.com/aptlogica/sereni-jwt-provider/actions/workflows/github-code-scanning/codeql/badge.svg" alt="CodeQL"></a>
<a href="https://sonarcloud.io/dashboard?id=aptlogica_sereni-jwt-provider"><img src="https://sonarcloud.io/api/project_badges/measure?project=aptlogica_sereni-jwt-provider&metric=alert_status" alt="Quality Gate"></a>
<a href="https://sonarcloud.io/dashboard?id=aptlogica_sereni-jwt-provider"><img src="https://sonarcloud.io/api/project_badges/measure?project=aptlogica_sereni-jwt-provider&metric=coverage" alt="Coverage"></a>
<a href="https://sonarcloud.io/dashboard?id=aptlogica_sereni-jwt-provider"><img src="https://sonarcloud.io/api/project_badges/measure?project=aptlogica_sereni-jwt-provider&metric=security_rating" alt="Security"></a>
</p>

<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
</p>

## Overview

**sereni-jwt-provider** is an open-source authentication service for generating, verifying, and managing JSON Web Tokens (JWT). It enables secure, token-based authentication for APIs, microservices, and backend applications. Designed for scalability and ease of integration, it helps protect endpoints, validate requests, and ensure only authorized access to your systems.

## Key Features

- **JWT Token Management**: Secure issuance, validation, and revocation of JWT tokens
- **Automated Key Rotation**: Configurable key rotation with seamless token transition
- **Multi-Tenant Support**: Isolated authentication contexts for different applications
- **Advanced Security**: RSA/ECDSA signing, token blacklisting, and security headers
- **Comprehensive Monitoring**: Detailed authentication metrics and audit logging
- **Token-Based Authentication**: JWT auth API with JWT provider capabilities for authentication microservice deployment
- **Cloud-Native Architecture**: Kubernetes-ready with horizontal scaling support

### Token Revocation

This service supports revoking specific tokens via an HTTP API endpoint. Example:

```bash
# Revoke a specific token
curl -X POST http://localhost:8081/api/v1/auth/revoke \
    -H 'Authorization: Bearer <admin-token>' \
    -H 'Content-Type: application/json' \
    -d '{"token": "<token-to-revoke>"}'
```

Revoked tokens are stored in an in-memory blacklist by default. On service restart the in-memory blacklist is cleared — for persistent revocation across restarts configure a Redis backend via `REDIS_URL` (see `ENV` configuration below).

### Multi-Tenant Context

`sereni-jwt-provider` supports isolated authentication contexts for multiple tenants. Recommended ways to pass a tenant ID:

- **HTTP header**: `X-Tenant-ID` — used for incoming API requests and routing to tenant-specific stores or configs.
- **JWT claim**: `tid` (tenant id) — included in issued tokens so downstream services can enforce tenant-scoped authorization.
- **Fallback / config**: a default tenant can be configured via environment variables for single-tenant deployments.

When both header and token claim are present, the service validates they match; otherwise the request is rejected.

## Architecture
- Go 1.26.2, idiomatic design
- Modular, testable codebase

## Installation
```sh
go get github.com/aptlogica/sereni-jwt-provider
```

## Configuration
See `.env.example` for environment variables and configuration options.

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/aptlogica/sereni-jwt-provider/pkg/client"
    "github.com/aptlogica/sereni-jwt-provider/pkg/config"
    "github.com/aptlogica/sereni-jwt-provider/pkg/types"
)

func main() {
    // Initialize configuration
    cfg := config.New()
    cfg.JWTSecret = "your-secret-key"
    cfg.TokenExpiry = "24h"
    cfg.RefreshExpiry = "7d"
    
    // Create JWT provider
    provider, err := client.New(cfg)
    if err != nil {
        log.Fatal("Failed to create provider:", err)
    }
    
    // Generate token for user
    claims := &types.Claims{
        UserID: "user123",
        Email:  "user@example.com",
        Roles:  []string{"user", "admin"},
    }
    
    ctx := context.Background()
    tokens, err := provider.GenerateTokens(ctx, claims)
    if err != nil {
        log.Fatal("Failed to generate tokens:", err)
    }
    
    log.Printf("Access Token: %s", tokens.AccessToken)
    log.Printf("Refresh Token: %s", tokens.RefreshToken)
}
```

## Development

### Local Setup
```bash
# Clone the repository
git clone https://github.com/aptlogica/sereni-jwt-provider.git
cd sereni-jwt-provider

# Install dependencies
go mod download

# Set up environment
cp .env.example .env
# Configure your JWT settings in .env

# Generate RSA keys for JWT signing
openssl genrsa -out private_key.pem 2048
openssl rsa -in private_key.pem -pubout -out public_key.pem

# Start development server
go run ./cmd/server
```

### Environment Configuration
```bash
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=24h
REFRESH_EXPIRY=7d
PRIVATE_KEY_PATH=./private_key.pem
PUBLIC_KEY_PATH=./public_key.pem
PORT=8080
LOG_LEVEL=debug
```

### Key Management
```bash
# Generate new RSA key pair
make generate-keys

# Rotate keys (zero-downtime)
make rotate-keys
```

### Zero-Downtime Key Rotation

`make rotate-keys` generates a new RSA key pair and triggers a graceful transition:

- New tokens are signed with the new key immediately.
- Old tokens signed with the previous key remain valid until their natural expiry.
- After all old tokens expire, the old key is removed from the active key set.

This approach ensures no users are logged out during rotation. If you rely on very short token lifetimes, schedule rotations carefully to avoid overlapping key removal before tokens expire.

If you need manual control or inspection, the key material is stored in the configured keys directory; use your deployment automation to snapshot or distribute the public keys to dependent services.

## Testing
- Run `go test ./...` to execute unit tests

## Repository Topics (recommended)

The GitHub repository's topics should reflect this is a Go microservice (not Node.js/TypeScript). Recommended topics:

- `go`, `golang`, `microservice`, `apache-2-0`, `open-source`, `jwt`, `auth`

You can update topics using the GitHub CLI, for example:

```bash
# Example: replace topics via GitHub CLI
gh repo edit aptlogica/sereni-jwt-provider --add-topic go golang microservice apache-2-0 open-source jwt auth --remove-topic nodejs typescript
```

## SereniBase Ecosystem

This service is part of the SereniBase platform. The core platform repository `sereni-base` relies on `sereni-jwt-provider` for centralized authentication. See the platform root here:

- https://github.com/aptlogica/sereni-base
## Security
See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License
Apache License 2.0. Copyright (c) 2026 Aptlogica Technologies.


