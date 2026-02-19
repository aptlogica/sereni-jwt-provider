# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Install Swagger CLI tool
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/server/main.go -o docs

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# ==============================================================================
# Production Stage
# ==============================================================================
FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app


# Copy built binary from builder
COPY --from=builder /app/main ./main

# Copy Swagger documentation
COPY --from=builder /app/docs ./docs

# Create non-root user for security
RUN adduser -D -s /bin/sh jwtprovider && \
    chown -R jwtprovider:jwtprovider /app

# Switch to non-root user
USER jwtprovider

# Expose application port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]
