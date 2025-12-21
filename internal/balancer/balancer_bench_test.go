package balancer

import (
	"LoadBalancer/internal/backend"
	"fmt"
	"testing"
)

var benchmarkCounts = []int{10, 100, 1000, 10000, 100000, 1000000, 10000000}

func BenchmarkRoundRobin_Pick(b *testing.B) {
	for _, count := range benchmarkCounts {
		b.Run(fmt.Sprintf("%d", count), func(b *testing.B) {
			benchmarkBalancer_Pick(b, "round_robin", count)
		})
	}
}

func BenchmarkLeastConnections_Pick(b *testing.B) {
	for _, count := range benchmarkCounts {
		b.Run(fmt.Sprintf("%d", count), func(b *testing.B) {
			benchmarkBalancer_Pick(b, "least_connections", count)
		})
	}
}

func BenchmarkWeighted_Pick(b *testing.B) {
	for _, count := range benchmarkCounts {
		b.Run(fmt.Sprintf("%d", count), func(b *testing.B) {
			benchmarkBalancer_Pick(b, "weighted", count)
		})
	}
}

func BenchmarkIPHash_Pick(b *testing.B) {
	for _, count := range benchmarkCounts {
		b.Run(fmt.Sprintf("%d", count), func(b *testing.B) {
			benchmarkBalancer_Pick(b, "ip_hash", count)
		})
	}
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
