package sdk

import (
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ResourceEventHandler alias for cache.ResourceEventHandler
type ResourceEventHandler = cache.ResourceEventHandler

// ResourceEventHandlerFuncs alias for cache.ResourceEventHandlerFuncs
type ResourceEventHandlerFuncs = cache.ResourceEventHandlerFuncs

// Watcher interface defines methods for watching resources
type Watcher interface {
	// Start starts all registered informers
	Start(stopCh <-chan struct{})

	// WaitForCacheSync waits for caches to sync
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	// WatchPods registers a handler for Pod events
	WatchPods(namespace string, handler ResourceEventHandler) error

	// WatchDeployments registers a handler for Deployment events
	WatchDeployments(namespace string, handler ResourceEventHandler) error

	// WatchNodes registers a handler for Node events
	WatchNodes(handler ResourceEventHandler) error

	// WatchServices registers a handler for Service events
	WatchServices(namespace string, handler ResourceEventHandler) error

	// WatchEvents registers a handler for Kubernetes Events
	WatchEvents(namespace string, handler ResourceEventHandler) error
}

type watcher struct {
	clientset kubernetes.Interface
	factory   informers.SharedInformerFactory
}

// NewWatcher creates a new Watcher instance
// defaultResync is the default resync period for all informers
func NewWatcher(clientset kubernetes.Interface, defaultResync time.Duration) Watcher {
	return &watcher{
		clientset: clientset,
		factory:   informers.NewSharedInformerFactory(clientset, defaultResync),
	}
}

func (w *watcher) Start(stopCh <-chan struct{}) {
	w.factory.Start(stopCh)
}

func (w *watcher) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	return w.factory.WaitForCacheSync(stopCh)
}

// filterHandler wraps a ResourceEventHandler to filter events by namespace
type filterHandler struct {
	namespace string
	handler   ResourceEventHandler
}

func (f *filterHandler) OnAdd(obj interface{}, isInInitialList bool) {
	if f.matches(obj) {
		f.handler.OnAdd(obj, isInInitialList)
	}
}

func (f *filterHandler) OnUpdate(oldObj, newObj interface{}) {
	if f.matches(newObj) {
		f.handler.OnUpdate(oldObj, newObj)
	}
}

func (f *filterHandler) OnDelete(obj interface{}) {
	if f.matches(obj) {
		f.handler.OnDelete(obj)
	}
}

func (f *filterHandler) matches(obj interface{}) bool {
	if f.namespace == "" {
		return true
	}
	
	object, err := meta.Accessor(obj)
	if err != nil {
		// Try to handle DeletedFinalStateUnknown
		if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			if object, err = meta.Accessor(tombstone.Obj); err == nil {
				return object.GetNamespace() == f.namespace
			}
		}
		return false
	}
	return object.GetNamespace() == f.namespace
}

func (w *watcher) addHandler(informer cache.SharedIndexInformer, namespace string, handler ResourceEventHandler) error {
	var h ResourceEventHandler = handler
	if namespace != "" {
		h = &filterHandler{namespace: namespace, handler: handler}
	}
	_, err := informer.AddEventHandler(h)
	return err
}

func (w *watcher) WatchPods(namespace string, handler ResourceEventHandler) error {
	return w.addHandler(w.factory.Core().V1().Pods().Informer(), namespace, handler)
}

func (w *watcher) WatchDeployments(namespace string, handler ResourceEventHandler) error {
	return w.addHandler(w.factory.Apps().V1().Deployments().Informer(), namespace, handler)
}

func (w *watcher) WatchNodes(handler ResourceEventHandler) error {
	// Nodes are cluster-scoped, so namespace is ignored
	return w.addHandler(w.factory.Core().V1().Nodes().Informer(), "", handler)
}

func (w *watcher) WatchServices(namespace string, handler ResourceEventHandler) error {
	return w.addHandler(w.factory.Core().V1().Services().Informer(), namespace, handler)
}

func (w *watcher) WatchEvents(namespace string, handler ResourceEventHandler) error {
	return w.addHandler(w.factory.Core().V1().Events().Informer(), namespace, handler)
}
