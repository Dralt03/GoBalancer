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
	"LoadBalancer/internal/logging"
	"LoadBalancer/internal/proxy"

	"go.uber.org/zap"
)

func main() {
	//Load configuration file
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := logging.Init("info", "console"); err != nil {
		log.Fatalf("Failed to initialise logger: %v", err)
	}

	defer logging.L().Sync()

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
		logging.L().Fatal("Invalid load balancing algorithm", zap.String("algorithm", cfg.Algorithm))
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
		logging.L().Fatal("Failed to create proxy", zap.Error(err))
	}

	go func() {
		if err := pxy.Start(); err != nil {
			logging.L().Fatal("Failed to start proxy", zap.Error(err))
		}
	}()

	//Graceful Shutdown
	sigC := make(chan os.Signal, 1) //Buffered channel to avoid missing signals
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	logging.L().Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pxy.Stop(ctx); err != nil {
		logging.L().Error("Failed to stop proxy", zap.Error(err))
	}
	logging.L().Info("Load Balanced exited cleanly.")
}
