package kubernetes

import (
	"LoadBalancer/pkg/discovery"
	"context"
	"fmt"
)

type kubernetesDiscover struct {
	namespace string
	service   string
}

func NewKubernetesDiscover(namespace, service string) *kubernetesDiscover {
	return &kubernetesDiscover{
		namespace: namespace,
		service:   service,
	}
}

func (k *kubernetesDiscover) Run(ctx context.Context, events chan<- discovery.Event) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fmt.Println("watch endpoint slices")
		}
	}
}
