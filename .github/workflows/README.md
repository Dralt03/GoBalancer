# GitHub Actions Workflows

This directory contains automated CI/CD workflows for GoBalancer.

## Workflows

### 1. CI (`ci.yml`)

**Triggers:** Push and Pull Requests to `main`/`master` branches

**Jobs:**
- **Test** - Runs tests on Go 1.21, 1.22, and 1.23
  - Downloads dependencies
  - Runs `go vet`
  - Runs tests with race detector
  - Generates coverage reports
  - Uploads coverage to Codecov
  
- **Build** - Builds the binary for Linux amd64
  - Creates binary artifact
  - Uploads for 7-day retention
  
- **Cross-Platform Build** - Builds for all platforms
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64, arm64)
  - Uploads artifacts for each platform
  
- **Lint** - Runs golangci-lint
  - Comprehensive linting checks
  - 5-minute timeout
  
- **Docker** - Builds Docker image
  - Uses BuildKit cache
  - Verifies image builds successfully

### 2. Release (`release.yml`)

**Triggers:** Push of version tags (e.g., `v1.0.0`)

**Jobs:**
- **Release** - Creates GitHub release
  - Runs tests before building
  - Builds binaries for all platforms
  - Creates `.tar.gz` (Linux/macOS) and `.zip` (Windows) archives
  - Generates SHA256 checksums
  - Creates changelog from git commits
  - Uploads release assets to GitHub
  
- **Docker Release** - Publishes Docker images
  - Builds multi-platform images (amd64, arm64)
  - Pushes to GitHub Container Registry (ghcr.io)
  - Tags: `latest`, `vX.Y.Z`, `vX.Y`, `vX`

### 3. Go Format Check (`gofmt.yml`)

**Triggers:** Push and Pull Requests to `main`/`master` branches

**Jobs:**
- **gofmt** - Verifies code formatting
  - Checks all `.go` files are formatted
  - Fails if any files need formatting

## Usage

### Running CI on Pull Requests

CI runs automatically on all pull requests. Ensure:
- All tests pass
- Code is properly formatted (`gofmt -w .`)
- Linting checks pass

### Docker Images

Published Docker images are available at:
```bash
docker pull ghcr.io/dralt03/gobalancer:latest
docker pull ghcr.io/dralt03/gobalancer:v1.0.0
```

## Secrets Required

The workflows use the following secrets:
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `CODECOV_TOKEN` - (Optional) For Codecov integration

## Badge Status

Add these badges to your README:

```markdown
[![CI](https://github.com/Dralt03/GoBalancer/actions/workflows/ci.yml/badge.svg)](https://github.com/Dralt03/GoBalancer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dralt03/GoBalancer)](https://goreportcard.com/report/github.com/Dralt03/GoBalancer)
```

## Local Testing

Before pushing, you can run the same checks locally:

```bash
# Format check
gofmt -l .

# Tests
go test -race ./...

# Vet
go vet ./...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Build
go build -o gobalancer main.go
```
