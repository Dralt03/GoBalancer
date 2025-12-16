package docker

import (
	"LoadBalancer/internal/logging"
	"LoadBalancer/pkg/discovery"
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type dockerDiscover struct {
	client *client.Client
	// Map container ID to Address (host:port) to handle removals
	containers map[string]string
}

func NewDockerDiscover() *dockerDiscover {
	// Initialize client with environment variables and version negotiation
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logging.L().Error("Failed to create docker client", zap.Error(err))
		return nil
	}
	return &dockerDiscover{
		client:     cli,
		containers: make(map[string]string),
	}
}

func (d *dockerDiscover) Run(ctx context.Context, apiEvents chan<- discovery.Event) error {
	if d.client == nil {
		return fmt.Errorf("docker client not initialized")
	}
	
	logging.L().Info("Docker discovery started")

	// 1. Initial Sync: List existing containers
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "lb.enable=true")),
	})
	if err != nil {
		return fmt.Errorf("failed to list existing containers: %w", err)
	}

	for _, c := range containers {
		d.handleAdd(ctx, c.ID, apiEvents)
	}

	// 2. Event Watch Loop
	eventFilters := filters.NewArgs()
	eventFilters.Add("type", "container")
	eventFilters.Add("label", "lb.enable=true")
	eventFilters.Add("event", "start")
	eventFilters.Add("event", "die")
	eventFilters.Add("event", "pause")
	eventFilters.Add("event", "unpause")
	eventFilters.Add("event", "kill")

	msgs, errs := d.client.Events(ctx, events.ListOptions{
		Filters: eventFilters,
	})

	for {
		select {
		case err := <-errs:
			return fmt.Errorf("docker event error: %w", err)
		case <-ctx.Done():
			return nil
		case msg := <-msgs:
			switch msg.Action {
			case "start", "unpause":
				d.handleAdd(ctx, msg.Actor.ID, apiEvents)
			case "die", "pause", "kill":
				d.handleRemove(msg.Actor.ID, apiEvents)
			}
		}
	}
}

func (d *dockerDiscover) handleAdd(ctx context.Context, containerID string, apiEvents chan<- discovery.Event) {
	// Inspect to get details (IP, Labels)
	info, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		logging.L().Error("Failed to inspect container", zap.String("id", containerID), zap.Error(err))
		return
	}

	if info.State.Paused || !info.State.Running {
		return
	}

	// Extract IP
	var ip string
	// Try default network first, iterate others if needed
	for _, network := range info.NetworkSettings.Networks {
		ip = network.IPAddress
		if ip != "" {
			break
		}
	}

	if ip == "" {
		logging.L().Warn("Container has no IP", zap.String("id", containerID), zap.String("name", info.Name))
		return
	}

	// Extract Port
	port := "80"
	if p, ok := info.Config.Labels["lb.port"]; ok {
		port = p
	}

	address := fmt.Sprintf("%s:%s", ip, port)

	// Extract Weight
	weight := int64(1)
	if wStr, ok := info.Config.Labels["lb.weight"]; ok {
		if w, err := strconv.ParseInt(wStr, 10, 64); err == nil {
			weight = w
		}
	}

	// Update State
	d.containers[containerID] = address

	logging.L().Info("Discovered backend", zap.String("address", address), zap.Int64("weight", weight))
	apiEvents <- discovery.Event{
		Type:    discovery.BackendAdd,
		Address: address,
		Weight:  weight,
	}
}

func (d *dockerDiscover) handleRemove(containerID string, apiEvents chan<- discovery.Event) {
	addr, exists := d.containers[containerID]
	if !exists {
		return
	}

	logging.L().Info("Removing backend", zap.String("address", addr))
	apiEvents <- discovery.Event{
		Type:    discovery.BackendRemove,
		Address: addr,
	}
	
	delete(d.containers, containerID)
}
