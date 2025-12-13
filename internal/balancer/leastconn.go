package balancer

import (
	"LoadBalancer/internal/backend"
	"errors"
)

type LeastConnections struct {
	pool *backend.Pool
}

func NewLeastConnectionsBalancer(pool *backend.Pool) *LeastConnections {
	return &LeastConnections{
		pool: pool,
	}
}

func (lc *LeastConnections) Pick(_ string) (*backend.Backend, error) {
	backends := lc.pool.AliveSnapshot()
	n := len(backends)
	if n == 0 {
		return nil, errors.New("no alive backends")
	}

	selected := backends[0]
	minCount := selected.ConnCount()
	for _, b := range backends {
		currCount := b.ConnCount()
		if currCount < minCount {
			selected = b
			minCount = currCount
		}
	}
	selected.IncConn()
	return selected, nil
}
