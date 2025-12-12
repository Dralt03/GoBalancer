package backend

import (
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	Address string
	Weight  int
	mu      sync.RWMutex

	alive               int32 // 1=UP 0=DOWN
	connCount           int64
	consecutiveFailures int32
	consecutiveSuccess  int32
	lastFailed          time.Time
	lastSuccess         time.Time
}

func NewBackend(address string, weight int) *Backend {
	b := &Backend{
		Address: address,
		Weight:  weight,
	}

	// Backend is considered healthy by default until marked by health checker
	atomic.StoreInt32(&b.alive, 1)
	return b
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (b *Backend) MarkAlive() {
	atomic.StoreInt32(&b.alive, 1)
	atomic.StoreInt32(&b.consecutiveFailures, 0)
	atomic.AddInt32(&b.consecutiveSuccess, 1)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastSuccess = time.Now()
}

func (b *Backend) MarkDead() {
	atomic.StoreInt32(&b.alive, 0)
	atomic.StoreInt32(&b.consecutiveSuccess, 0)
	atomic.AddInt32(&b.consecutiveFailures, 1)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastFailed = time.Now()
}

func (b *Backend) IncConn() {
	atomic.AddInt64(&b.connCount, 1)
}

func (b *Backend) DecConn() {
	atomic.AddInt64(&b.connCount, -1)
}

func (b *Backend) ConnCount() int64 {
	return atomic.LoadInt64(&b.connCount)
}

func (b *Backend) GetLastSuccess() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastSuccess
}

func (b *Backend) GetLastFailed() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastFailed
}

func (b *Backend) AddFailures() int32 {
	return atomic.AddInt32(&b.consecutiveFailures, 1)
}

func (b *Backend) ResetFailures() {
	atomic.StoreInt32(&b.consecutiveFailures, 0)
}

func (b *Backend) AddSuccess() int32 {
	return atomic.AddInt32(&b.consecutiveSuccess, 1)
}

func (b *Backend) ResetSuccess() {
	atomic.StoreInt32(&b.consecutiveSuccess, 0)
}
