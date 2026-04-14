# Sereni JWT Provider Makefile
# Standard commands for Go development workflow

.PHONY: help setup test test-race test-coverage coverage coverage-func lint build clean install dev run docker release

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := sereni-jwt-provider
VERSION := $(shell git describe --tags --always --dirty 2>nul || echo dev)
BUILD_TIME := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'" 2>nul || date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go env GOVERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GO_VERSION)"
COVER_PROFILE := coverage.out
COVER_HTML := coverage.html

##@ Help
help: ## Display this help message
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available Targets:"
	@echo "  setup              - Install development dependencies"
	@echo "  dev                - Run in development mode"
	@echo "  run                - Run the application"
	@echo "  install            - Install the binary to GOPATH/bin"
	@echo "  test               - Run all tests"
	@echo "  test-race          - Run tests with race detection"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-benchmark     - Run benchmark tests"
	@echo "  coverage           - Alias for test-coverage"
	@echo "  coverage-func      - Show coverage by function"
	@echo "  lint               - Run linter"
	@echo "  lint-fix           - Run linter with auto-fix"
	@echo "  security           - Run security scanner"
	@echo "  vet                - Run go vet"
	@echo "  fmt                - Format code"
	@echo "  check              - Run all quality checks"
	@echo "  build              - Build the binary"
	@echo "  build-all          - Build for all platforms"
	@echo "  clean              - Clean build artifacts"
	@echo ""

##@ Development
setup: ## Install development dependencies
	@echo "🔧 Installing development dependencies..."
	@go mod download
	@go mod tidy
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.54.2)
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	@echo "✅ Development setup complete"

dev: ## Run in development mode with hot reload
	@echo "🚀 Starting development server..."
	@go run main.go

run: dev ## Run the application

install: ## Install the binary to $GOPATH/bin
	@echo "📦 Installing $(BINARY_NAME) to $$GOPATH/bin..."
	@go install $(LDFLAGS) ./...
	@echo "✅ Installation complete"

##@ Testing
test: ## Run all tests
	@echo "Running tests..."
	cmd /c "go test -v -race -coverprofile=$(COVER_PROFILE) -covermode=atomic -coverpkg=./internal/... ./tests/..."

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -race -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	cmd /c "go test -v -race -coverprofile=$(COVER_PROFILE) -covermode=atomic -coverpkg=./internal/... ./tests/..."
	cmd /c "go tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)"
	cmd /c "go tool cover -func=$(COVER_PROFILE)"
	@echo "Coverage report generated: $(COVER_HTML)"

coverage: test-coverage ## Alias for test-coverage

coverage-func: ## Show coverage by function
	@go tool cover -func=$(COVER_PROFILE)
test-benchmark: ## Run benchmark tests
	@echo "⚡ Running benchmark tests..."
	@go test -bench=. -benchmem ./...

##@ Code Quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run --timeout=5m

lint-fix: ## Run linter and fix issues automatically
	@echo "🔧 Running linter with auto-fix..."
	@golangci-lint run --fix --timeout=5m

security: ## Run security scanner
	@echo "🔒 Running security scanner..."
	@gosec ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...

check: lint vet security test ## Run all quality checks

##@ Build
build: ## Build the binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "🔨 Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✅ Multi-platform builds complete in dist/"

##@ Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker build -t $(BINARY_NAME):latest .
	@echo "✅ Docker image built: $(BINARY_NAME):$(VERSION)"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	@docker run --rm -p 8081:8081 $(BINARY_NAME):latest

docker-test: ## Test Docker image
	@echo "🧪 Testing Docker image..."
	@docker run --rm $(BINARY_NAME):latest --help

##@ Examples
examples: ## Run all examples
	@echo "📚 Running examples..."
	@cd examples/basic-auth && go run main.go &
	@sleep 2
	@cd examples/middleware && go run main.go &
	@echo "✅ Examples running. Check http://localhost:8080"

##@ Release
release-dry: ## Dry run of release process
	@echo "🚀 Dry run of release process..."
	@goreleaser release --snapshot --skip-publish --clean

release: ## Create a new release
	@echo "🚀 Creating release..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@goreleaser release --clean

##@ Maintenance  
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ dist/ $(COVER_PROFILE) $(COVER_HTML)
	@go clean -cache -modcache -i -r
	@echo "✅ Cleanup complete"

update-deps: ## Update dependencies
	@echo "📦 Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated"

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

##@ Documentation
docs-serve: ## Serve documentation locally
	@echo "📚 Starting documentation server..."
	@godoc -http=:6060
	@echo "✅ Documentation available at http://localhost:6060"

##@ CI/CD Helpers
ci-setup: setup ## Setup for CI environment
	@echo "🤖 Setting up CI environment..."

ci-test: test test-coverage lint security ## Run all CI tests
	@echo "🤖 CI tests complete"

ci-build: build docker-build ## Build artifacts for CI
	@echo "🤖 CI build complete"


