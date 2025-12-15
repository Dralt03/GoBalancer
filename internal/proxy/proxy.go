package proxy

import (
	"LoadBalancer/internal/config"
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Options struct {
	IOUring bool
	Timeout config.TimeoutCfg
}

type Proxy struct {
	listener net.Listener
	handler  *Handler

	wg       sync.WaitGroup
	stopOnce sync.Once

	ctx    context.Context
	cancel context.CancelFunc

	stopped int32
}

func NewProxy(address string, balancer Balancer, options Options) (*Proxy, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	h := NewHandler(balancer, options.Timeout)

	ctx, cancel := context.WithCancel(context.Background())
	return &Proxy{
		listener: listener,
		handler:  h,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (p *Proxy) Start() error {
	log.Printf("Proxy Listening on port: %s\n", p.listener.Addr().String())

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			if atomic.LoadInt32(&p.stopped) == 1 {
				return nil
			}

			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			log.Printf("Accept error: %v\n", err)
			continue
		}

		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	defer p.wg.Done()
	p.handler.Handle(p.ctx, conn)
}

func (p *Proxy) Stop(ctx context.Context) error {
	var err error

	p.stopOnce.Do(func() {
		atomic.StoreInt32(&p.stopped, 1)
		err = p.listener.Close()
		p.cancel()
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Proxy Closed")
		return err
	case <-ctx.Done():
		log.Printf("Context Done")
		return ctx.Err()
	}
}
