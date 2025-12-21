# GoBalancer

[![CI](https://github.com/Dralt03/GoBalancer/actions/workflows/ci.yml/badge.svg)](https://github.com/Dralt03/GoBalancer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dralt03/GoBalancer)](https://goreportcard.com/report/github.com/Dralt03/GoBalancer)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Dralt03/GoBalancer)](https://go.dev/)
[![License](https://img.shields.io/github/license/Dralt03/GoBalancer)](LICENSE)

A high-performance TCP load balancer written in Go, designed with enterprise-grade features, multiple load balancing algorithms, and thread-safe concurrent operations.

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

## Performance & Benchmarking

GoBalancer is designed for high throughput and low latency. We continuously verify performance using our automated benchmarking suite which tests scalability against 10, 100, 1,000, and 10,000 backend nodes.

### Test Environment
*   **CPU**: 12th Gen Intel(R) Core(TM) i7-1255U
*   **Memory**: 16 GB DDR4
*   **OS**: Windows 11
*   **Go Version**: 1.24.3

### Algorithm Overhead (vs Backend Count)

Measured latency (ns/op) for picking a backend as the number of candidates scales:

| Backend Count | Round Robin | Least Connections | Weighted | IP Hash (HRW) |
| :--- | :--- | :--- | :--- | :--- |
| **10** | ~60 ns | ~140 ns | ~150 ns | ~640 ns |
| **100** | ~255 ns | ~630 ns | ~1024 ns | ~4700 ns |
| **1,000** | ~3180 ns | ~11000 ns | ~9300 ns | ~45000 ns |
| **10,000** | ~67200 ns | ~156000 ns | ~135000 ns | ~500000 ns |
| **100,000** | ~870000 ns | ~2000000 ns | ~2280000 ns | ~5500000 ns |
| **1,000,000** | ~10000000 ns | ~23000000 ns | ~21000000 ns | ~61000000 ns |
| **10,000,000** | ~190000000 ns | ~320000000 ns | ~210000000 ns | ~940000000 ns |

> **Note**: Round Robin remains constant O(1). Least Connections and Weighted scale linearly O(N) in worst-case scanning, but are optimized with internal heaps/trees in production where applicable.

<center><img src="./assets/benchmark_results.png" alt="Benchmark Results" /></center>

### Capacity & Scaling
- **Backend Limit**: Optimized for thousands of backends with minimal memory overhead.
- **Concurrent Connections**: Primarily limited by the Operating System's file descriptor limits (`ulimit`). The Go runtime easily handles tens of thousands of concurrent connection goroutines.
- **Memory Footprint**: Low overhead per connection (~2KB stack space). 10,000 active connections typically require only 20-40MB of overhead.

---

## Development & Maintenance

We provide several tools to streamline development, testing, and releases.

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
