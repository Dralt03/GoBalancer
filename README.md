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

## REST API Server

The REST API **runs automatically** on port `8081` when you start the load balancer. It provides runtime management capabilities.

**Endpoints:**
- `GET /health` - API health check
- `GET /backends` - List all backends with status
- `POST /backends` - Add a new backend
- `GET /backends/{address}` - Get specific backend details
- `PUT /backends/{address}` - Update backend weight
- `DELETE /backends/{address}` - Remove a backend

**Example:**
```bash
# List all backends
curl http://localhost:8081/backends

# Add a backend
curl -X POST http://localhost:8081/backends \
  -H "Content-Type: application/json" \
  -d '{"address": "192.168.1.100:8080", "weight": 2}'
```

## Service Discovery

GoBalancer supports three discovery modes:

### 1. Static Discovery (Default)

Backends are manually defined in `config.yaml`:

```yaml
discovery:
  type: "static"

backends:
  - address: "10.0.0.4:3000"
    weight: 1
  - address: "10.0.0.5:3000"
    weight: 2
```

### 2. Docker Discovery

**Automatically discovers containers** with the `lb.enable=true` label.

**Configuration:**
```yaml
discovery:
  type: "docker"
```

**Docker Compose Example:**
```yaml
services:
  gobalancer:
    image: gobalancer
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro  # Required!
    ports:
      - "8080:8080"
      - "8081:8081"

  backend1:
    image: nginx:alpine
    labels:
      - "lb.enable=true"      # Required for discovery
      - "lb.port=80"          # Optional, defaults to 80
      - "lb.weight=2"         # Optional, defaults to 1
```

**How it works:**
- Watches Docker events (container start/stop)
- Automatically adds containers with `lb.enable=true` label
- Uses container's internal IP address
- Removes backends when containers stop

### 3. Kubernetes Discovery

**Automatically discovers pods** from a Kubernetes service.

**Configuration:**
```yaml
discovery:
  type: "kubernetes"
  kubernetes:
    namespace: "default"
    service: "my-backend-service"
```

**Requirements:**
- RBAC permissions to watch EndpointSlices (see `deployments/kubernetes/rbac.yaml`)
- Backend pods must be part of a Kubernetes Service

**How it works:**
- Watches EndpointSlices for the specified service
- Automatically adds/removes backends when pods scale
- Only adds pods that are "ready" (respects readiness probes)
- Uses pod IP addresses and service port

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
