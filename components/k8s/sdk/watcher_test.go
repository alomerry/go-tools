package sdk

import (
	"context"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWatchPods(t *testing.T) {
	// Create fake clientset
	clientset := fake.NewSimpleClientset()

	// Create watcher
	watcher := NewWatcher(clientset, time.Minute)

	// Create channel to signal event received
	eventCh := make(chan string, 1)

	// Register handler
	handler := ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			eventCh <- pod.Name
		},
	}

	err := watcher.WatchPods("default", handler)
	if err != nil {
		t.Fatalf("Failed to watch pods: %v", err)
	}

	// Start watcher
	stopCh := make(chan struct{})
	defer close(stopCh)
	watcher.Start(stopCh)
	watcher.WaitForCacheSync(stopCh)

	// Create a pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}
	_, err = clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// Wait for event
	select {
	case name := <-eventCh:
		if name != "test-pod" {
			t.Errorf("Expected pod name 'test-pod', got '%s'", name)
		}
	case <-time.After(time.Second * 2):
		t.Error("Timeout waiting for pod event")
	}
}

func TestWatchNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	watcher := NewWatcher(clientset, time.Minute)

	var wg sync.WaitGroup
	wg.Add(1)

	// Watch only "ns1"
	handler := ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if pod.Namespace != "ns1" {
				t.Errorf("Received event for pod in namespace '%s', expected 'ns1'", pod.Namespace)
			}
			wg.Done()
		},
	}

	err := watcher.WatchPods("ns1", handler)
	if err != nil {
		t.Fatalf("Failed to watch pods: %v", err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	watcher.Start(stopCh)
	watcher.WaitForCacheSync(stopCh)

	// Create pod in "ns2" - should be ignored
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-ns2",
			Namespace: "ns2",
		},
	}
	_, err = clientset.CoreV1().Pods("ns2").Create(context.Background(), pod2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod in ns2: %v", err)
	}

	// Create pod in "ns1" - should be detected
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-ns1",
			Namespace: "ns1",
		},
	}
	_, err = clientset.CoreV1().Pods("ns1").Create(context.Background(), pod1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod in ns1: %v", err)
	}

	// Wait for event with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second * 2):
		t.Error("Timeout waiting for filtered pod event")
	}
}

func TestWatchEvents(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	watcher := NewWatcher(clientset, time.Minute)

	eventCh := make(chan string, 1)

	handler := ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event := obj.(*corev1.Event)
			eventCh <- event.Reason
		},
	}

	err := watcher.WatchEvents("default", handler)
	if err != nil {
		t.Fatalf("Failed to watch events: %v", err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	watcher.Start(stopCh)
	watcher.WaitForCacheSync(stopCh)

	// Create an Event
	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		Reason:  "Created",
		Message: "Pod created successfully",
	}
	_, err = clientset.CoreV1().Events("default").Create(context.Background(), k8sEvent, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	select {
	case reason := <-eventCh:
		if reason != "Created" {
			t.Errorf("Expected event reason 'Created', got '%s'", reason)
		}
	case <-time.After(time.Second * 2):
		t.Error("Timeout waiting for k8s event")
	}
}
