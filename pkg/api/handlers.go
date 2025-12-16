package api

import (
	"LoadBalancer/internal/backend"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	pool *backend.Pool
}

func NewHandler(pool *backend.Pool) *Handler {
	return &Handler{
		pool: pool,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) GetBackends(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		backends := h.pool.GetBackends()
		response := make([]Backend, 0, len(backends))
		for _, b := range backends {
			response = append(response, Backend{
				Address:   b.Address,
				Weight:    b.GetWeight(),
				Alive:     b.IsAlive(),
				ConnCount: b.ConnCount(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		var req AddBackendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := h.pool.AddBackend(req.Address, req.Weight)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Backend{
			Address:   b.Address,
			Weight:    b.GetWeight(),
			Alive:     b.IsAlive(),
			ConnCount: b.ConnCount(),
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) BackendByAddress(w http.ResponseWriter, r *http.Request) {
	address := strings.TrimPrefix(r.URL.Path, "/backends/")
	if address == "" {
		http.Error(w, "Backend address is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		b, err := h.pool.GetBackend(address)
		if err != nil {
			http.Error(w, "Backend not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Backend{
			Address:   b.Address,
			Weight:    b.GetWeight(),
			Alive:     b.IsAlive(),
			ConnCount: b.ConnCount(),
		})

	case http.MethodPut:
		var req UpdateWeightRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.pool.UpdateWeight(address, req.Weight); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		if !h.pool.RemoveBackend(address) {
			http.Error(w, "Backend not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
