package proxy

import (
	"LoadBalancer/internal/backend"
	"LoadBalancer/internal/config"
	"context"
	"log"
	"net"
	"time"
)

type Balancer interface {
	Pick(key string) (*backend.Backend, error)
}

type Handler struct {
	Balancer Balancer
	Timeouts config.TimeoutCfg
}

func NewHandler(balancer Balancer, timeouts config.TimeoutCfg) *Handler {
	return &Handler{
		Balancer: balancer,
		Timeouts: timeouts,
	}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	backend, err := h.Balancer.Pick(clientIP)
	if err != nil {
		log.Printf("failed to pick backend: %v", err)
		return
	}
	defer backend.DecConn()

	timeout := time.Duration(h.Timeouts.ConnectTimeout) * time.Second
	backendConn, err := net.DialTimeout("tcp", backend.Address, timeout)
	if err != nil {
		log.Printf("failed to connect to backend %s: %v", backend.Address, err)
		return
	}
	defer backendConn.Close()

	pipe(ctx, conn, backendConn)
}
