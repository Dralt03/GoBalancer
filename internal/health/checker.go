package health

import (
	"LoadBalancer/internal/backend"
	"LoadBalancer/internal/config"
	"context"
	"log"
	"net"
	"sync"
	"time"
)

type Checker struct {
	pool   *backend.Pool
	config config.HealthCfg
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(pool *backend.Pool, config config.HealthCfg) *Checker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Checker{
		pool:   pool,
		config: config,
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}
}

func (c *Checker) checkBackend(backend *backend.Backend) {
	defer c.wg.Done()

	dialer := net.Dialer{
		Timeout: time.Duration(c.config.TimeoutSec) * time.Second,
	}

	conn, err := dialer.DialContext(c.ctx, "tcp", backend.Address)
	if err != nil {
		failures := backend.AddFailures()
		if failures >= int32(c.config.Retries) {
			backend.MarkDead()
		}
		return
	}

	conn.Close()
	backend.ResetFailures()
	if !backend.IsAlive() {
		backend.MarkAlive()
	}
}

func (c *Checker) runOnce() {
	backends := c.pool.GetBackends()
	for _, backend := range backends {
		c.wg.Add(1)
		go c.checkBackend(backend)
	}
}

func (c *Checker) Start() {
	log.Println("Health Checker Started")

	ticker := time.NewTicker(time.Duration(c.config.IntervalSec) * time.Second)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.runOnce()
			case <-c.ctx.Done():
				log.Println("Health Checker Stopped")
				return
			}
		}
	}()
}
