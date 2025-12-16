package docker

import (
	"LoadBalancer/pkg/discovery"
	"context"
	"fmt"
)

type dockerDiscover struct {
}

func NewDockerDiscover() *dockerDiscover {
	return &dockerDiscover{}
}

func (d *dockerDiscover) Run(ctx context.Context, events chan<- discovery.Event) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fmt.Println("docker discover")
		}
	}
}
