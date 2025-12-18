package balancer

import (
	"LoadBalancer/internal/backend"
	"fmt"
	"testing"
)

func BenchmarkRoundRobin_Pick_1000(b *testing.B) {
	benchmarkBalancer_Pick(b, "round_robin", 1000)
}

func BenchmarkLeastConnections_Pick_1000(b *testing.B) {
	benchmarkBalancer_Pick(b, "least_connections", 1000)
}

func BenchmarkWeighted_Pick_1000(b *testing.B) {
	benchmarkBalancer_Pick(b, "weighted", 1000)
}

func BenchmarkIPHash_Pick_1000(b *testing.B) {
	benchmarkBalancer_Pick(b, "ip_hash", 1000)
}

func benchmarkBalancer_Pick(b *testing.B, algo string, count int) {
	pool := backend.NewPool()
	for i := 0; i < count; i++ {
		_, _ = pool.AddBackend(fmt.Sprintf("10.0.0.%d:8080", i), 1)
	}

	var lb Balancer
	switch algo {
	case "round_robin":
		lb = NewRoundRobinBalancer(pool)
	case "least_connections":
		lb = NewLeastConnectionsBalancer(pool)
	case "weighted":
		lb = NewWeightedBalancer(pool)
	case "ip_hash":
		lb = NewIPHashBalancer(pool)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = lb.Pick("192.168.1.1")
	}
}
