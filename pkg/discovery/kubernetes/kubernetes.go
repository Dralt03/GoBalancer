package kubernetes

import (
	"LoadBalancer/internal/logging"
	"LoadBalancer/pkg/discovery"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesDiscover struct {
	clientset kubernetes.Interface
	namespace string
	service   string
}

func NewKubernetesDiscover(namespace, service string) *kubernetesDiscover {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to local kubeconfig
		home, _ := os.UserHomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logging.L().Error("Failed to load kubeconfig", zap.Error(err))
			return nil
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logging.L().Error("Failed to create kubernetes client", zap.Error(err))
		return nil
	}

	return &kubernetesDiscover{
		clientset: clientset,
		namespace: namespace,
		service:   service,
	}
}

func (k *kubernetesDiscover) Run(ctx context.Context, eventsChan chan<- discovery.Event) error {
	if k.clientset == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	logging.L().Info("Starting Kubernetes discovery", zap.String("namespace", k.namespace), zap.String("service", k.service))

	// Watch EndpointSlices for the service
	labelSelector := fmt.Sprintf("kubernetes.io/service-name=%s", k.service)
	watcher, err := k.clientset.DiscoveryV1().EndpointSlices(k.namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				logging.L().Warn("Watcher channel closed, restarting")
				// Simple restart mechanism: return error to let caller handle restart or reimplement loop here
				// For now, returning error to allow main loop (if any) to backoff and retry
				return fmt.Errorf("watcher channel closed")
			}

			if event.Type == watch.Error {
				logging.L().Error("Watcher error", zap.Any("object", event.Object))
				continue
			}

			slice, ok := event.Object.(*discoveryv1.EndpointSlice)
			if !ok {
				continue
			}

			k.handleEndpointSlice(slice, event.Type, eventsChan)
		}
	}
}

func (k *kubernetesDiscover) handleEndpointSlice(slice *discoveryv1.EndpointSlice, eventType watch.EventType, eventsChan chan<- discovery.Event) {
	for _, endpoint := range slice.Endpoints {
		// Skip if not ready
		ready := true
		if endpoint.Conditions.Ready != nil {
			ready = *endpoint.Conditions.Ready
		}
		if !ready {
			continue
		}

		// Use the first IP for now
		if len(endpoint.Addresses) == 0 {
			continue
		}
		ip := endpoint.Addresses[0]

		// Find port (assume HTTP/80 or use first defined port)
		var port int32 = 80
		if len(slice.Ports) > 0 && slice.Ports[0].Port != nil {
			port = *slice.Ports[0].Port
		}

		address := fmt.Sprintf("%s:%d", ip, port)

		// Determine action
		var discoType discovery.EventType
		switch eventType {
		case watch.Added, watch.Modified:
			discoType = discovery.BackendAdd
		case watch.Deleted:
			discoType = discovery.BackendRemove
		default:
			continue
		}

		// Weight defaults to 1 for now
		weight := int64(1)

		logging.L().Info("Kubernetes discovery event",
			zap.String("type", string(eventType)),
			zap.String("address", address),
		)

		eventsChan <- discovery.Event{
			Type:    discoType,
			Address: address,
			Weight:  weight,
		}
	}
}
