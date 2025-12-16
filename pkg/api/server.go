package api

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	http *http.Server
}

func New(address string, handler http.Handler) *Server {
	return &Server{
		http: &http.Server{
			Addr:    address,
			Handler: handler,
			ReadTimeout: 5*time.Second,
			WriteTimeout: 5*time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}