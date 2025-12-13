package balancer

import (
	"LoadBalancer/internal/backend"
	"errors"
	"hash/fnv"
)

// Uses Highest Random Weight Hashing
// For each score = hash(ClientIP + backendAddress) and maximum score is picked

type IPHash struct {
	pool *backend.Pool
}

func NewIPHashBalancer(pool *backend.Pool) *IPHash {
	return &IPHash{
		pool: pool,
	}
}

func hrwHash(key, backend string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	h.Write([]byte(backend))
	return h.Sum64()
}

func (ip *IPHash) Pick(clientIP string) (*backend.Backend, error) {
	backends := ip.pool.AliveSnapshot()
	n := len(backends)
	if n == 0 {
		return nil, errors.New("no alive backends")
	}

	var selected *backend.Backend
	var maxScore uint64

	for _, b := range backends {
		score := hrwHash(clientIP, b.Address)
		if selected == nil || score > maxScore {
			selected = b
			maxScore = score
		}
	}

	selected.IncConn()
	return selected, nil
}
