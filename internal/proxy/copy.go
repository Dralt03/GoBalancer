package proxy

import (
	"io"
	"net"
	"sync"
)

func pipe(a, b net.Conn){
	var wg sync.WaitGroup
	wg.Add(2)
	go func(){
		defer wg.Done()
		_, err := io.Copy(b, a)
		if err != nil {
			b.Close()
		}
		closeWrite(b)
	}()

	go func(){
		defer wg.Done()
		_, err := io.Copy(a, b)
		if err != nil {
			a.Close()
		}
		closeWrite(a)
	}()

	wg.Wait()
}

func closeWrite(conn net.Conn){
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.CloseWrite()
	}
}