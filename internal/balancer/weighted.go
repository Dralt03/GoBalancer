package balancer

import (
	"LoadBalancer/internal/backend"
	"errors"
	"math"
)

type Weighted struct {
	pool *backend.Pool
}

func NewWeightedBalancer(pool *backend.Pool) *Weighted {
	return &Weighted{
		pool: pool,
	}
}

// Picking a backend based on the minimum score for the backend achieved using the formula:
// score = (connections + 1) / weight
func (w *Weighted) Pick(_ string) (*backend.Backend, error) {
	backends := w.pool.AliveSnapshot()
	n := len(backends)
	if n == 0 {
		return nil, errors.New("no alive backends")
	}

	var selected *backend.Backend
	minScore := math.MaxFloat64

	for _, b := range backends {
		weight := b.GetWeight()
		if weight <= 0 {
			continue
		}

		score := float64(b.ConnCount()+1) / float64(weight)

		if score < minScore {
			minScore = score
			selected = b
		}
	}

	if selected == nil {
		return nil, errors.New("no backend selected")
	}

	selected.IncConn()
	return selected, nil
}
