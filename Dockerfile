# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.24-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Install Swagger CLI tool (pinned to v1.16.4 - commit 0b9e347c196710ea155a147782bf51707a600c2c)
RUN git clone https://github.com/swaggo/swag.git /tmp/swag && \
    cd /tmp/swag && \
    git checkout 0b9e347c196710ea155a147782bf51707a600c2c && \
    go install ./cmd/swag && \
    rm -rf /tmp/swag

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/server/main.go -o docs

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -installsuffix cgo -o main ./cmd/server

# ==============================================================================
# Production Stage
# ==============================================================================
FROM alpine:3.20@sha256:a4f4213abb84c497377b8544c81b3564f313746700372ec4fe84653e4fb03805

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app


# Copy built binary from builder
COPY --from=builder /app/main ./main

# Copy Swagger documentation
COPY --from=builder /app/docs ./docs

# Create non-root user for security
RUN adduser -D -s /bin/sh jwtprovider && chown -R jwtprovider:jwtprovider /app

# Switch to non-root user
USER jwtprovider

# Expose application port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD curl -f http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]
