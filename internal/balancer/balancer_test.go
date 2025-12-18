package balancer

import (
	"LoadBalancer/internal/backend"
	"fmt"
	"testing"
)

func TestLeastConnections(t *testing.T) {
	pool := backend.NewPool()
	b1, _ := pool.AddBackend("10.0.0.1:8080", 1)
	b2, _ := pool.AddBackend("10.0.0.2:8080", 1)

	lb := NewLeastConnectionsBalancer(pool)

	// b1 has 5 connections, b2 has 2 connections
	for i := 0; i < 5; i++ {
		b1.IncConn()
	}
	for i := 0; i < 2; i++ {
		b2.IncConn()
	}

	// Should pick b2 as it has fewer connections
	picked, err := lb.Pick("")
	if err != nil {
		t.Fatalf("Failed to pick: %v", err)
	}
	if picked.Address != "10.0.0.2:8080" {
		t.Errorf("Expected 10.0.0.2:8080, got %s", picked.Address)
	}

	// Now b2 has 3 connections. Pick again, still b2.
	picked, _ = lb.Pick("")
	if picked.Address != "10.0.0.2:8080" {
		t.Errorf("Expected 10.0.0.2:8080, got %s", picked.Address)
	}

	// Now b2 has 4 connections. Pick again, still b2.
	picked, _ = lb.Pick("")
	if picked.Address != "10.0.0.2:8080" {
		t.Errorf("Expected 10.0.0.2:8080, got %s", picked.Address)
	}

	// Now b2 has 5 connections. Next pick could be either b1 or b2.
	// Since we iterate, it depends on order but logically it's correct.
}

func TestWeighted(t *testing.T) {
	pool := backend.NewPool()
	// b1: weight 10, b2: weight 2
	_, _ = pool.AddBackend("10.0.0.1:8080", 10)
	_, _ = pool.AddBackend("10.0.0.2:8080", 2)

	lb := NewWeightedBalancer(pool)

	// Initial score for b1: (0+1)/10 = 0.1
	// Initial score for b2: (0+1)/2 = 0.5
	// Should pick b1
	picked, err := lb.Pick("")
	if err != nil {
		t.Fatalf("Failed to pick: %v", err)
	}
	if picked.Address != "10.0.0.1:8080" {
		t.Errorf("Expected 10.0.0.1:8080, got %s", picked.Address)
	}

	// After 5 picks, b1 should still be preferred mostly
	for i := 0; i < 4; i++ {
		picked, _ = lb.Pick("")
		if picked.Address != "10.0.0.1:8080" {
			t.Errorf("Expected 10.0.0.1:8080 at pick %d, got %s", i+2, picked.Address)
		}
	}

	// Pick 6: b1 has 5 connections -> score (5+1)/10 = 0.6
	// b2 has 0 connections -> score (0+1)/2 = 0.5
	// Should pick b2
	picked, _ = lb.Pick("")
	if picked.Address != "10.0.0.2:8080" {
		t.Errorf("Expected 10.0.0.2:8080, got %s", picked.Address)
	}
}

func TestIPHash(t *testing.T) {
	pool := backend.NewPool()
	pool.AddBackend("10.0.0.1:8080", 1)
	pool.AddBackend("10.0.0.2:8080", 1)
	pool.AddBackend("10.0.0.3:8080", 1)

	lb := NewIPHashBalancer(pool)

	// Same IP should result in same backend
	picked1, _ := lb.Pick("192.168.1.1")
	picked2, _ := lb.Pick("192.168.1.1")
	if picked1.Address != picked2.Address {
		t.Errorf("Session persistence failed: expected %s, got %s", picked1.Address, picked2.Address)
	}

	// Different IP might result in different backend
	picked3, _ := lb.Pick("192.168.1.2")
	fmt.Printf("IP 1.1 -> %s, IP 1.2 -> %s\n", picked1.Address, picked3.Address)
}
