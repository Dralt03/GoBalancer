package main

import (
	"context"
	"log"
	"net/http"
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
	"LoadBalancer/pkg/api"
	"LoadBalancer/pkg/discovery"
	"LoadBalancer/pkg/discovery/docker"
	"LoadBalancer/pkg/discovery/kubernetes"

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

	defer func() { _ = logging.L().Sync() }()

	//Initialise backend pool
	pool := backend.NewPool()

	for _, b := range cfg.Backends {
		if _, err := pool.AddBackend(b.Address, b.Weight); err != nil {
			logging.L().Error("Failed to add initial backend", zap.String("address", b.Address), zap.Error(err))
		}
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

	// App context for long-running services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup if main exits early

	events := make(chan discovery.Event, 128)

	// Initialize discovery services
	switch cfg.Discovery.Type {
	case "docker":
		logging.L().Info("Using Docker discovery")
		dockerDiscover := docker.NewDockerDiscover()
		if dockerDiscover != nil {
			go func() {
				if err := dockerDiscover.Run(ctx, events); err != nil {
					logging.L().Error("Docker discovery failed", zap.Error(err))
				}
			}()
		}
	case "kubernetes":
		logging.L().Info("Using Kubernetes discovery")
		k8sDiscover := kubernetes.NewKubernetesDiscover(
			cfg.Discovery.Kubernetes.Namespace,
			cfg.Discovery.Kubernetes.Service,
		)
		go func() {
			if err := k8sDiscover.Run(ctx, events); err != nil {
				logging.L().Error("Kubernetes discovery failed", zap.Error(err))
			}
		}()
	case "static":
		logging.L().Info("Using static discovery")
	default:
		logging.L().Warn("Unknown discovery type, defaulting to static", zap.String("type", cfg.Discovery.Type))
	}

	registry := backend.NewRegistry(pool)

	go func() {
		for e := range events {
			registry.Apply(e)
		}
	}()

	go func() {
		if err := pxy.Start(); err != nil {
			logging.L().Fatal("Failed to start proxy", zap.Error(err))
		}
	}()

	apiHandler := api.NewHandler(pool)
	apiRouter := api.Routes(apiHandler)
	apiServer := api.New(":8081", apiRouter)

	go func() {
		logging.L().Info("API Server Listening", zap.String("address", ":8081"))
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			logging.L().Fatal("Failed to start API server", zap.Error(err))
		}
	}()

	//Graceful Shutdown
	sigC := make(chan os.Signal, 1) //Buffered channel to avoid missing signals
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	logging.L().Info("Shutting down gracefully...")

	// Cancel long-running services
	cancel()

	// Shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := pxy.Stop(shutdownCtx); err != nil {
		logging.L().Error("Failed to stop proxy", zap.Error(err))
	}

	if err := apiServer.Stop(shutdownCtx); err != nil {
		logging.L().Error("Failed to stop API server", zap.Error(err))
	}

	logging.L().Info("Load Balanced exited cleanly.")
}
