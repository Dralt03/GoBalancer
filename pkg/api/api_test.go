package api

import (
	"LoadBalancer/internal/backend"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	pool := backend.NewPool()
	h := NewHandler(pool)
	server := httptest.NewServer(Routes(h))
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetBackends(t *testing.T) {
	pool := backend.NewPool()
	h := NewHandler(pool)
	server := httptest.NewServer(Routes(h))
	defer server.Close()

	// Initially empty
	resp, err := http.Get(server.URL + "/backends")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var backends []Backend
	if err := json.NewDecoder(resp.Body).Decode(&backends); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(backends) != 0 {
		t.Errorf("Expected 0 backends, got %d", len(backends))
	}

	// Add one
	_, _ = pool.AddBackend("10.0.0.1:8080", 1)

	resp, err = http.Get(server.URL + "/backends")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&backends); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(backends) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(backends))
	}
	if backends[0].Address != "10.0.0.1:8080" {
		t.Errorf("Expected address 10.0.0.1:8080, got %s", backends[0].Address)
	}
}

func TestAddBackend(t *testing.T) {
	pool := backend.NewPool()
	h := NewHandler(pool)
	server := httptest.NewServer(Routes(h))
	defer server.Close()

	reqBody := AddBackendRequest{
		Address: "10.0.0.2:8080",
		Weight:  10,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/backends", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Verify it was added
	if !pool.HasBackend("10.0.0.2:8080") {
		t.Error("Backend was not added to pool")
	}

	// Test duplicate
	resp, err = http.Post(server.URL+"/backends", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate, got %d", resp.StatusCode)
	}
}

func TestBackendByAddress(t *testing.T) {
	pool := backend.NewPool()
	_, _ = pool.AddBackend("10.0.0.3:8080", 5)

	h := NewHandler(pool)
	server := httptest.NewServer(Routes(h))
	defer server.Close()

	// GET
	resp, err := http.Get(server.URL + "/backends/10.0.0.3:8080")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var b Backend
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if b.Weight != 5 {
		t.Errorf("Expected weight 5, got %d", b.Weight)
	}

	// PUT
	updateReq := UpdateWeightRequest{Weight: 20}
	updateBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, server.URL+"/backends/10.0.0.3:8080", bytes.NewReader(updateBody))

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	updatedDetails, _ := pool.GetBackend("10.0.0.3:8080")
	if updatedDetails.GetWeight() != 20 {
		t.Errorf("Expected updated weight 20, got %d", updatedDetails.GetWeight())
	}

	// DELETE
	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/backends/10.0.0.3:8080", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	if pool.HasBackend("10.0.0.3:8080") {
		t.Error("Backend was not removed from pool")
	}

	// Test 404
	resp, err = http.Get(server.URL + "/backends/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
