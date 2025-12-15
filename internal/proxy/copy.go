package proxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

func pipe(ctx context.Context, a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to signal that legitimate copying is done
	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			// Shutdown signal received, force close connections
			a.SetDeadline(time.Now())
			b.SetDeadline(time.Now())
		case <-done:
			// Normal completion, exit to avoid leak
			return
		}
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(b, a)
		if err != nil {
			b.Close()
		}
		closeWrite(b)
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(a, b)
		if err != nil {
			a.Close()
		}
		closeWrite(a)
	}()

	wg.Wait()
	close(done)
}

func closeWrite(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.CloseWrite()
	}
}
