package balancer

import "LoadBalancer/internal/backend"

type Balancer interface {
	Pick() (*backend.Backend, error)
}