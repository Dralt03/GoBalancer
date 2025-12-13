package backend

import (
	"errors"
	"sync"
)

type Pool struct {
	mu       sync.RWMutex
	backends []*Backend
	index    map[string]*Backend
}

func NewPool() *Pool {
	return &Pool{
		backends: make([]*Backend, 0, 8),
		index:    make(map[string]*Backend),
	}
}

func (p *Pool) AddBackend(address string, weight int64) (*Backend, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[address]; ok {
		return nil, errors.New("backend already exists")
	}

	b := NewBackend(address, weight)
	p.backends = append(p.backends, b)
	p.index[address] = b
	return b, nil
}

func (p *Pool) RemoveBackend(address string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[address]; !ok{
		return false
	}

	delete(p.index, address)
	newBackends := make([]*Backend, 0, len(p.backends)-1)
	for _, b := range p.backends{
		if b.Address == address{
			continue
		}
		newBackends = append(newBackends, b)
	}
	p.backends = newBackends
	return true
}

func (p *Pool) GetBackends() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Backend, len(p.backends))
	copy(out, p.backends)
	return out
}

func (p *Pool) GetBackend(address string) (*Backend, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if b, ok := p.index[address]; ok {
		return b, nil
	}
	return nil, errors.New("backend not found")
}

func (p *Pool) HasBackend(address string) bool{
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.index[address]
	return ok
}

func (p *Pool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.backends)
}

func (p *Pool) UpdateWeight(address string, weight int64) error {
	p.mu.RLock()
	b, ok := p.index[address]
	p.mu.RUnlock()

	if !ok {
		return errors.New("backend not found")
	}

	b.SetWeight(weight)
	return nil
}

func (p *Pool) MarkAlive(address string) error {
	p.mu.RLock()
	b, ok := p.index[address]
	p.mu.RUnlock()

	if !ok {
		return errors.New("backend not found")
	}

	b.MarkAlive()
	return nil
}

func (p *Pool) MarkDead(address string) error {
	p.mu.RLock()
	b, ok := p.index[address]
	p.mu.RUnlock()

	if !ok {
		return errors.New("backend not found")
	}

	b.MarkDead()
	return nil
}

func (p *Pool) AliveSnapshot() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.IsAlive() {
			out = append(out, b)
		}
	}
	return out
}