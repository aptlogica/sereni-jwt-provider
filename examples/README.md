# Sereni JWT Provider Examples

This directory contains practical examples demonstrating how to use the Sereni JWT Provider in various scenarios.

## Examples Overview

| Example | Description | Complexity |
|---------|-------------|------------|
| [Basic Authentication](./basic-auth/) | Simple JWT token generation and validation | Beginner |
| [Refresh Tokens](./refresh-tokens/) | Token refresh and rotation | Intermediate |
| [Middleware Integration](./middleware/) | HTTP middleware for route protection | Intermediate |
| [Multi-Tenant JWT](./multi-tenant/) | JWT with tenant-specific claims | Advanced |
| [OAuth Integration](./oauth/) | OAuth2 flow with JWT tokens | Advanced |

## Quick Start

Choose an example that matches your use case and follow the README in each directory.

### Basic Usage

```go
go run basic-auth/main.go
```

### With Custom Configuration

```go
go run middleware/main.go
```

## Prerequisites

- Go 1.21+
- Port 8080 available (for examples with HTTP server)

## Common Configuration

Most examples use these environment variables:

```bash
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h
JWT_ISSUER=sereni-jwt-provider
```