package balancer

import "LoadBalancer/internal/backend"

type Balancer interface {
	Pick(key string) (*backend.Backend, error)
}