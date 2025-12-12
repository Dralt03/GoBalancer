package backend

import "sync"

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
