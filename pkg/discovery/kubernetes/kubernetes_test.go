package kubernetes

import (
	"LoadBalancer/pkg/discovery"
	"context"
	"testing"
	"time"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestKubernetesDiscover_Run(t *testing.T) {
	// 1. Setup Fake Client
	clientset := fake.NewSimpleClientset()

	k := &kubernetesDiscover{
		clientset: clientset,
		namespace: "default",
		service:   "my-service",
	}

	eventsChan := make(chan discovery.Event, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Start Discovery in background
	go func() {
		if err := k.Run(ctx, eventsChan); err != nil {
			// Run returns error on context cancel, which is expected
			return
		}
	}()

	// Allow watcher to start
	time.Sleep(100 * time.Millisecond)

	// 3. Create an EndpointSlice
	slice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-service-slice-1",
			Namespace: "default",
			Labels: map[string]string{
				"kubernetes.io/service-name": "my-service",
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: ptr.To(true),
				},
			},
		},
		Ports: []discoveryv1.EndpointPort{
			{
				Port: ptr.To(int32(8080)),
			},
		},
	}

	_, err := clientset.DiscoveryV1().EndpointSlices("default").Create(ctx, slice, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create EndpointSlice: %v", err)
	}

	// 4. Verify Event
	select {
	case event := <-eventsChan:
		if event.Type != discovery.BackendAdd {
			t.Errorf("Expected BackendAdd, got %v", event.Type)
		}
		expectedAddr := "10.0.0.1:8080"
		if event.Address != expectedAddr {
			t.Errorf("Expected address %s, got %s", expectedAddr, event.Address)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// 5. Update EndpointSlice (Modify IP)
	slice.Endpoints[0].Addresses = []string{"10.0.0.2"}
	_, err = clientset.DiscoveryV1().EndpointSlices("default").Update(ctx, slice, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update EndpointSlice: %v", err)
	}

	// Verify Update Event
	select {
	case event := <-eventsChan:
		// Client-go modifies are seen as Add events in our logic (upsert)
		if event.Type != discovery.BackendAdd {
			t.Errorf("Expected BackendAdd (Update), got %v", event.Type)
		}
		expectedAddr := "10.0.0.2:8080"
		if event.Address != expectedAddr {
			t.Errorf("Expected address %s, got %s", expectedAddr, event.Address)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for update event")
	}

	// 6. Delete EndpointSlice
	err = clientset.DiscoveryV1().EndpointSlices("default").Delete(ctx, slice.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete EndpointSlice: %v", err)
	}

	// Verify Delete Event
	// Note: Delete events in EndpointSlices are tricky because the slice is gone.
	// Our logic handles watch.Deleted.
	select {
	case event := <-eventsChan:
		if event.Type != discovery.BackendRemove {
			t.Errorf("Expected BackendRemove, got %v", event.Type)
		}
		// Based on our implementation, we should get the address back from the deleted object
		// provided the watch event contains the last known state.
		expectedAddr := "10.0.0.2:8080"
		if event.Address != expectedAddr {
			t.Errorf("Expected address %s, got %s", expectedAddr, event.Address)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for delete event")
	}
}
