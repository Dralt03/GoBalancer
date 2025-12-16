package backend

import (
	"LoadBalancer/pkg/discovery"
)

type registry struct {
	pool *Pool
}

func NewRegistry(pool *Pool) *registry {
	return &registry{
		pool: pool,
	}
}

func (r *registry) Apply(event discovery.Event) {
	switch event.Type {
	case discovery.BackendAdd:
		r.pool.AddBackend(event.Address, event.Weight)
	case discovery.BackendRemove:
		r.pool.RemoveBackend(event.Address)
	}
}
