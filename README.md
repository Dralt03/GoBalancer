# GoBalancer

A high-performance HTTP/TCP load balancer written in Go, designed with enterprise-grade features, modern load balancing algorithms, and thread-safe concurrent operations.

## Features

### Implemented

- **Multiple Load Balancing Algorithms**
  - Round Robin - Evenly distribute requests across backends
  - Least Connections - Route to backend with fewest active connections
  - Weighted - Distribute based on backend capacity weights

- **Thread-Safe Backend Management**
  - Lock-free atomic operations for performance-critical counters
  - RWMutex protection for timestamp tracking
  - Safe concurrent access from multiple goroutines
  - Hybrid synchronization strategy for optimal performance

- **Health Checking**
  - Active health probes (TCP/HTTP)
  - Automatic backend failover
  - Graceful backend recovery

- **Configuration Management**
  - YAML/TOML configuration support
  - Flexible backend definitions
  - Configurable health check parameters

- **Production Ready**
  - Graceful shutdown handling
  - Signal handling (SIGINT, SIGTERM)
  - Docker containerization
  - Kubernetes deployment configurations

### In Progress

- **Advanced I/O**
  - Platform-optimized I/O multiplexing (epoll, kqueue, io_uring)
  - Connection pooling and buffer management
  - Zero-copy optimizations where possible

- **Observability**
  - Structured logging with custom formatters
  - Prometheus metrics endpoint (planned)

## Project Structure

```
GoBalancer/
├── internal/                    # Private application code
│   ├── backend/                # Backend management
│   │   ├── node.go            # Thread-safe backend node
│   │   └── pool.go            # Backend pool management
│   │
│   ├── balancer/               # Load balancing algorithms
│   │   ├── roundrobin.go      # Round-robin implementation
│   │   ├── leastconn.go       # Least connections algorithm
│   │   └── weighted.go        # Weighted distribution
│   │
│   ├── health/                 # Health checking
│   │   └── checker.go         # Health probe implementations
│   │
│   ├── config/                 # Configuration loading
│   │   └── config.go          # YAML/TOML config parser
│   │
│   ├── io/                     # High-performance I/O (in progress)
│   │   ├── poller.go          # Event polling interface
│   │   ├── epoll.go           # Linux epoll implementation
│   │   ├── kqueue.go          # BSD/macOS kqueue implementation
│   │   ├── uring.go           # io_uring support (Linux 5.1+)
│   │   └── buffers.go         # Buffer pool management
│   │
│   ├── logging/                # Logging infrastructure
│   │   ├── logger.go          # Logger implementation
│   │   └── formatter.go       # Log formatters
│   │
│   └── proxy/                  # Proxy implementation (in progress)
│
├── pkg/                         # Public API
│   └── api/                    # API client
│       └── client.go          # Load balancer API client
│
├── config/                      # Configuration files
│   └── config.yaml             # Sample configuration
│
├── deployments/                 # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yaml
│   └── kubernetes/
│       ├── deployment.yaml    # K8s deployment
│       ├── service.yaml       # K8s service
│       ├── configmap.yaml     # Configuration
│       └── rbac.yaml          # Role-based access control
│
├── scripts/                     # Build and utility scripts
│   ├── build.sh               # Build script
│   ├── release.sh             # Release automation
│   └── test.sh                # Test runner
│
├── test/                        # Test files and fixtures
│
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
└── LICENSE                      # MIT License
```

## Technology Stack

- **Language**: Go 1.21+
- **I/O Multiplexing**: Platform-specific (epoll/kqueue/io_uring)
- **Deployment**: Docker, Kubernetes

## Prerequisites

- Go 1.21 or higher
- Docker (optional, for containerized deployment)
- Kubernetes cluster (optional, for K8s deployment)

## Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/Dralt03/GoBalancer.git
cd GoBalancer

# Install dependencies
go mod download

# Build the project
go build -o gobalancer main.go
```

### Running

```bash
# Run directly
./gobalancer

# Or with Go
go run main.go
```

## Docker Deployment

```bash
# Build Docker image
docker build -f deployments/docker/Dockerfile -t gobalancer:latest .

# Run with docker-compose
docker-compose -f deployments/docker/docker-compose.yaml up
```

## Kubernetes Deployment

```bash
# Apply all Kubernetes configurations
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -l app=gobalancer
kubectl get svc gobalancer
```

## Testing

```bash
# Run tests
./scripts/test.sh

# Or manually
go test ./...

# Run with coverage
go test -cover ./...
```

## Performance

GoBalancer is designed for high throughput and low latency:

- **Platform-optimized I/O**: Uses epoll (Linux), kqueue (BSD/macOS), or io_uring for maximum performance
- **Connection pooling**: Reuses connections to backends
- **Zero-copy**: Minimizes data copying where possible
- **Concurrent health checks**: Non-blocking health probe execution
- **Lock-free counters**: Atomic operations for frequently-accessed fields

## Thread Safety & Concurrency

GoBalancer uses a **hybrid synchronization strategy** for optimal performance under high concurrency:

### Backend Node Synchronization

The `Backend` struct is fully thread-safe and can be accessed concurrently from multiple goroutines:

**Atomic Operations** (Lock-Free):
- `alive` (int32) - Backend health status
- `connCount` (int64) - Active connection counter
- `consecutiveFailures` (int32) - Failure tracking
- `consecutiveSuccess` (int32) - Success tracking

These fields use `sync/atomic` operations for lock-free reads and writes, providing optimal performance for high-frequency operations like connection tracking and health status checks.

**RWMutex Protection**:
- `lastSuccess` (time.Time) - Last successful health check timestamp
- `lastFailed` (time.Time) - Last failed health check timestamp

Time fields use `sync.RWMutex` for safe concurrent access:
- Multiple goroutines can read timestamps concurrently using `RLock()`
- Write operations use exclusive `Lock()` to prevent data races

### Design Rationale

This hybrid approach balances performance and safety:
-  **Lock-free fast path**: Frequently accessed counters avoid mutex overhead
-  **Safe timestamp access**: Struct-type time.Time requires mutex protection
-  **Read-optimized**: RWMutex allows concurrent readers for timestamp queries
-  **Race-free**: All concurrent access is properly synchronized

## Roadmap

### Completed

- [x] Implement core load balancing algorithms (Round Robin, Least Connections, Weighted)
- [x] Add health checking system
- [x] Configuration file support (YAML/TOML)
- [x] Thread-safe backend management
- [x] Graceful shutdown
- [x] Docker deployment
- [x] Kubernetes deployment manifests

### In Progress / Planned

- [ ] Complete platform-specific I/O optimizations (epoll, kqueue, io_uring)
- [ ] Finalize proxy implementation
- [ ] Prometheus metrics endpoint
- [ ] Structured logging system
- [ ] Rate limiting
- [ ] TLS/SSL termination
- [ ] WebSocket support
- [ ] gRPC load balancing
- [ ] Admin API for runtime configuration
- [ ] Hot reload capability
- [ ] Connection pooling optimizations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Dralt03**

⭐ Star this repository if you find it helpful!
