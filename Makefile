# GoBalancer Makefile
# Provides convenient targets for building, testing, and releasing

.PHONY: help build build-all run test test-quick bench fmt lint clean deps docker docker-run cross release

# Variables
BINARY_NAME=gobalancer
LOAD_TESTER=load-tester
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Default target
help:
	@echo "GoBalancer Makefile"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the load balancer for current platform"
	@echo "  build-all    Build both load balancer and load tester"
	@echo "  run          Build and run the load balancer"
	@echo "  test         Run tests with coverage"
	@echo "  test-quick   Run short tests without race detector"
	@echo "  bench        Run internal Go benchmarks"
	@echo "  fmt          Run go fmt and go mod tidy"
	@echo "  lint         Run linters (go vet + golangci-lint)"
	@echo "  clean        Clean build artifacts"
	@echo "  deps         Download and tidy dependencies"
	@echo "  docker       Build Docker image"
	@echo "  docker-run   Run load balancer in Docker"
	@echo "  cross        Cross-compile for all platforms"
	@echo "  release      Create a release (requires VERSION=vX.Y.Z)"
	@echo ""

# Build load balancer
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -trimpath $(LDFLAGS) -o bin/$(BINARY_NAME) main.go
	@echo "✓ Build complete: bin/$(BINARY_NAME)"

# Build load tester
build-tools:
	@echo "Building $(LOAD_TESTER)..."
	@mkdir -p bin
	@go build -trimpath -o bin/$(LOAD_TESTER) cmd/load_tester/main.go
	@echo "✓ Build complete: bin/$(LOAD_TESTER)"

build-all: build build-tools

# Build and run
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

# Run tests with coverage
test:
	@echo "Running tests..."
	@mkdir -p coverage
	@go test -race -v -coverprofile=coverage/coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage/coverage.out | grep total
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✓ Coverage report: coverage/coverage.html"

# Quick tests (short, no race detector)
test-quick:
	@echo "Running quick tests..."
	@go test -short ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@mkdir -p coverage
	@go test -v -bench=. -benchmem -run=^$$ ./... | tee coverage/bench.txt

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@go mod tidy
	@echo "✓ Formatting complete"

# Run linters
lint:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
		echo "✓ golangci-lint passed"; \
	else \
		echo "⚠ golangci-lint not found"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ release/ coverage/
	@rm -f $(BINARY_NAME) $(LOAD_TESTER) LoadBalancer lb *.exe
	@echo "✓ Clean complete"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Dependencies updated"

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -f deployments/docker/Dockerfile -t $(BINARY_NAME):latest .
	@echo "✓ Docker image built: $(BINARY_NAME):latest"

# Run in Docker
docker-run:
	@echo "Running $(BINARY_NAME) in Docker..."
	@docker run -p 8080:8080 -p 8081:8081 -v $$(pwd)/config:/app/config $(BINARY_NAME):latest

# Cross-compile for all platforms
cross:
	@echo "Cross-compiling for all platforms..."
	@if [ -f "./scripts/build.sh" ]; then \
		./scripts/build.sh --cross --release; \
	else \
		echo "Error: scripts/build.sh not found"; \
		exit 1; \
	fi
	@echo "✓ Cross-compilation complete"

# Create a release
release:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "Error: VERSION must be set (e.g., make release VERSION=v1.0.0)"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@if [ -f "./scripts/release.sh" ]; then \
		./scripts/release.sh $(VERSION); \
	else \
		echo "Error: scripts/release.sh not found"; \
		exit 1; \
	fi
	@echo "✓ Release $(VERSION) created"
