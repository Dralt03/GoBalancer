# GoBalancer

A high-performance HTTP/TCP load balancer written in Go, designed with enterprise-grade features, modern load balancing algorithms, and thread-safe concurrent operations.

## Features

- **Multiple Load Balancing Algorithms**
  - Round Robin - Evenly distribute requests across backends
  - Least Connections - Route to backend with fewest active connections
  - Weighted - Distribute based on backend capacity weights
  - IP Hash - Consistent routing based on client IP address

- **Service Discovery**
  - Static configuration
  - Docker container discovery
  - Kubernetes service discovery
  - Dynamic backend registration/deregistration

- **REST API Management**
  - Health check endpoint
  - List all backends with status
  - Add/remove backends dynamically
  - Update backend weights at runtime

- **Thread-Safe Backend Management**
  - Lock-free atomic operations for performance-critical counters
  - RWMutex protection for timestamp tracking
  - Safe concurrent access from multiple goroutines
  - Hybrid synchronization strategy for optimal performance

- **Health Checking**
  - Active health probes (TCP/HTTP)
  - Automatic backend failover
  - Graceful backend recovery

- **Production Ready**
  - Graceful shutdown handling
  - Docker and Kubernetes deployment
  - Structured logging with Zap

### In Progress

- Platform-optimized I/O multiplexing (epoll, kqueue, io_uring)
- Connection pooling and buffer management
- Zero-copy optimizations
- Prometheus metrics endpoint
- Distributed tracing support


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

## Running the Load Balancer

### Prerequisites

- Go 1.24 or higher
- Docker (optional, for containerized deployment)
- Kubernetes cluster (optional, for K8s deployment)

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

### Configuration

Edit `config/config.yaml`

### Run Locally

```bash
# Run directly
./gobalancer

# Or with Go
go run main.go
```

The load balancer will:
- Listen on port `8080` for incoming requests
- Expose REST API on port `8081` for management
- Automatically health check backends
- Route traffic based on the configured algorithm

### Run with Docker

```bash
# Build Docker image
docker build -f deployments/docker/Dockerfile -t gobalancer:latest .

# Run with docker-compose
docker-compose -f deployments/docker/docker-compose.yaml up
```

### Run on Kubernetes

```bash
# Apply all Kubernetes configurations
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -l app=gobalancer
kubectl get svc gobalancer
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Dralt03**

⭐ Star this repository if you find it helpful!
