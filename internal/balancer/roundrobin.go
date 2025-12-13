package balancer

import (
	"LoadBalancer/internal/backend"
	"errors"
	"sync/atomic"
)

type RoundRobin struct {
	pool *backend.Pool
	next uint64
}

func NewRoundRobinBalancer(pool *backend.Pool) *RoundRobin {
	return &RoundRobin{
		pool: pool,
	}
}

func (rr *RoundRobin) Pick(_ string) (*backend.Backend, error) {
	backends := rr.pool.AliveSnapshot()
	n := len(backends)
	if n == 0 {
		return nil, errors.New("no alive backends")
	}

	// Simple atomic increment and modulo
	next := atomic.AddUint64(&rr.next, 1)
	idx := (next - 1) % uint64(n)

	b := backends[idx]
	b.IncConn()
	return b, nil
}
