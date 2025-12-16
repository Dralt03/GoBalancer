package discovery

import "context"

type EventType int

const (
	BackendAdd EventType = iota
	BackendRemove
)

type Event struct {
	Type    EventType
	Address string
	Weight  int64
}

type Discover interface {
	Run(ctx context.Context, events chan<- Event) error
}
