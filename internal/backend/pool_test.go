package backend

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewPool(t *testing.T) {
	pool := NewPool()
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
	if pool.Len() != 0 {
		t.Errorf("Expected empty pool, got %d backends", pool.Len())
	}
}

func TestAddBackend(t *testing.T) {
	pool := NewPool()

	// Add first backend
	b, err := pool.AddBackend("10.0.0.1:8080", 1)
	if err != nil {
		t.Fatalf("Failed to add backend: %v", err)
	}
	if b == nil {
		t.Fatal("AddBackend returned nil backend")
	}
	if b.Address != "10.0.0.1:8080" {
		t.Errorf("Expected address 10.0.0.1:8080, got %s", b.Address)
	}
	if b.GetWeight() != 1 {
		t.Errorf("Expected weight 1, got %d", b.GetWeight())
	}

	// Verify it was added
	if pool.Len() != 1 {
		t.Errorf("Expected pool length 1, got %d", pool.Len())
	}
	if !pool.HasBackend("10.0.0.1:8080") {
		t.Error("Backend not found in pool")
	}

	// Try to add duplicate
	_, err = pool.AddBackend("10.0.0.1:8080", 2)
	if err == nil {
		t.Error("Expected error when adding duplicate backend")
	}
	if pool.Len() != 1 {
		t.Errorf("Expected pool length to remain 1, got %d", pool.Len())
	}
}

func TestRemoveBackend(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)
	_, _ = pool.AddBackend("10.0.0.2:8080", 1)

	// Remove existing backend
	removed := pool.RemoveBackend("10.0.0.1:8080")
	if !removed {
		t.Error("Failed to remove existing backend")
	}
	if pool.Len() != 1 {
		t.Errorf("Expected pool length 1 after removal, got %d", pool.Len())
	}
	if pool.HasBackend("10.0.0.1:8080") {
		t.Error("Backend still exists after removal")
	}

	// Try to remove non-existent backend
	removed = pool.RemoveBackend("nonexistent:8080")
	if removed {
		t.Error("Should not remove non-existent backend")
	}
	if pool.Len() != 1 {
		t.Errorf("Expected pool length to remain 1, got %d", pool.Len())
	}
}

func TestGetBackend(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 5)

	// Get existing backend
	b, err := pool.GetBackend("10.0.0.1:8080")
	if err != nil {
		t.Fatalf("Failed to get backend: %v", err)
	}
	if b.Address != "10.0.0.1:8080" {
		t.Errorf("Expected address 10.0.0.1:8080, got %s", b.Address)
	}
	if b.GetWeight() != 5 {
		t.Errorf("Expected weight 5, got %d", b.GetWeight())
	}

	// Get non-existent backend
	_, err = pool.GetBackend("nonexistent:8080")
	if err == nil {
		t.Error("Expected error when getting non-existent backend")
	}
}

func TestGetBackends(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)
	_, _ = pool.AddBackend("10.0.0.2:8080", 2)
	_, _ = pool.AddBackend("10.0.0.3:8080", 3)

	backends := pool.GetBackends()
	if len(backends) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(backends))
	}

	// Verify it's a copy (modifying shouldn't affect pool)
	backends[0] = nil
	if pool.Len() != 3 {
		t.Error("GetBackends should return a copy, not the original slice")
	}
}

func TestUpdateWeight(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)

	// Update existing backend
	err := pool.UpdateWeight("10.0.0.1:8080", 10)
	if err != nil {
		t.Fatalf("Failed to update weight: %v", err)
	}

	b, _ := pool.GetBackend("10.0.0.1:8080")
	if b.GetWeight() != 10 {
		t.Errorf("Expected weight 10, got %d", b.GetWeight())
	}

	// Update non-existent backend
	err = pool.UpdateWeight("nonexistent:8080", 5)
	if err == nil {
		t.Error("Expected error when updating non-existent backend")
	}
}

func TestMarkAliveAndDead(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)

	// Initially alive
	b, _ := pool.GetBackend("10.0.0.1:8080")
	if !b.IsAlive() {
		t.Error("Backend should be alive initially")
	}

	// Mark dead
	err := pool.MarkDead("10.0.0.1:8080")
	if err != nil {
		t.Fatalf("Failed to mark backend dead: %v", err)
	}
	if b.IsAlive() {
		t.Error("Backend should be dead after MarkDead")
	}

	// Mark alive
	err = pool.MarkAlive("10.0.0.1:8080")
	if err != nil {
		t.Fatalf("Failed to mark backend alive: %v", err)
	}
	if !b.IsAlive() {
		t.Error("Backend should be alive after MarkAlive")
	}

	// Mark non-existent backend
	err = pool.MarkDead("nonexistent:8080")
	if err == nil {
		t.Error("Expected error when marking non-existent backend")
	}
}

func TestAliveSnapshot(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)
	_, _ = pool.AddBackend("10.0.0.2:8080", 1)
	_, _ = pool.AddBackend("10.0.0.3:8080", 1)

	// All alive
	alive := pool.AliveSnapshot()
	if len(alive) != 3 {
		t.Errorf("Expected 3 alive backends, got %d", len(alive))
	}

	// Mark one dead
	err := pool.MarkDead("10.0.0.2:8080")
	if err != nil {
		t.Fatalf("Failed to mark backend dead: %v", err)
	}
	alive = pool.AliveSnapshot()
	if len(alive) != 2 {
		t.Errorf("Expected 2 alive backends, got %d", len(alive))
	}

	// Verify the dead one is not in snapshot
	for _, b := range alive {
		if b.Address == "10.0.0.2:8080" {
			t.Error("Dead backend should not be in alive snapshot")
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	pool := NewPool()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = pool.AddBackend(fmt.Sprintf("10.0.0.%d:8080", id), 1)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.GetBackends()
			pool.AliveSnapshot()
			pool.Len()
		}()
	}
	wg.Wait()

	// Concurrent updates
	backends := pool.GetBackends()
	for _, b := range backends {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			_ = pool.UpdateWeight(addr, 5)
			_ = pool.MarkDead(addr)
			_ = pool.MarkAlive(addr)
		}(b.Address)
	}
	wg.Wait()
}

func TestPoolEdgeCases(t *testing.T) {
	pool := NewPool()

	// Empty pool operations
	if pool.Len() != 0 {
		t.Error("New pool should be empty")
	}
	if pool.HasBackend("anything") {
		t.Error("Empty pool should not have any backends")
	}

	backends := pool.GetBackends()
	if len(backends) != 0 {
		t.Error("Empty pool should return empty slice")
	}

	alive := pool.AliveSnapshot()
	if len(alive) != 0 {
		t.Error("Empty pool should return empty alive snapshot")
	}

	// Operations on empty pool should fail gracefully
	_, err := pool.GetBackend("nonexistent")
	if err == nil {
		t.Error("GetBackend on empty pool should return error")
	}

	removed := pool.RemoveBackend("nonexistent")
	if removed {
		t.Error("RemoveBackend on empty pool should return false")
	}
}
