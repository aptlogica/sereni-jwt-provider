# sereni-jwt-provider - Secure JWT Authentication Service

> Enterprise-grade JWT authentication service and open source auth provider for secure application access. A comprehensive JWT auth server and identity provider offering advanced token management, key rotation, and seamless integration with modern authentication workflows.

[![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26.2-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Quality Gate Status](https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_sereni-jwt-provider_e93385ef-3917-48bf-b724-ca96963f99ce&metric=alert_status&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2)](https://sonar.aptlogica.com/dashboard?id=aptlogica_sereni-jwt-provider_e93385ef-3917-48bf-b724-ca96963f99ce)

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

## Testing
- Run `go test ./...` to execute unit tests

## Security
See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License
MIT License. Copyright (c) 2026 Aptlogica Technologies.


