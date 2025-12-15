package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"LoadBalancer/internal/backend"
	"LoadBalancer/internal/balancer"
	"LoadBalancer/internal/config"
	"LoadBalancer/internal/health"
	"LoadBalancer/internal/proxy"
)

func main() {
	//Load configuration file
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	//Initialise backend pool
	pool := backend.NewPool()

	for _, backend := range cfg.Backends {
		pool.AddBackend(backend.Address, backend.Weight)
	}

	var lb balancer.Balancer

	switch cfg.Algorithm {
	case "round_robin":
		lb = balancer.NewRoundRobinBalancer(pool)
	case "least_connections":
		lb = balancer.NewLeastConnectionsBalancer(pool)
	case "weighted":
		lb = balancer.NewWeightedBalancer(pool)
	case "ip_hash":
		lb = balancer.NewIPHashBalancer(pool)
	default:
		log.Fatalf("Invalid load balancing algorithm: %s", cfg.Algorithm)
	}

	hc := health.New(pool, cfg.HealthCheck)
	go hc.Start()

	pxy, err := proxy.NewProxy(
		cfg.ListenAddress,
		lb,
		proxy.Options{
			IOUring: cfg.UseIOUring,
			Timeout: cfg.Timeout,
		})
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	go func() {
		if err := pxy.Start(); err != nil {
			log.Fatalf("Failed to start proxy: %v", err)
		}
	}()

	//Graceful Shutdown
	sigC := make(chan os.Signal, 1) //Buffered channel to avoid missing signals
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pxy.Stop(ctx); err != nil {
		log.Printf("Failed to stop proxy: %v", err)
	}
	log.Println("Load Balanced exited cleanly.")
}
