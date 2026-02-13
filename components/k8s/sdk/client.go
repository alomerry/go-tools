package sdk

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset       *kubernetes.Clientset
	dynamicClient   *dynamic.DynamicClient
	discoveryClient *discovery.DiscoveryClient
	mapper          *restmapper.DeferredDiscoveryRESTMapper
	config          *rest.Config
}

// NewClient creates a new Kubernetes client
// If kubeconfig is empty, it will try to use in-cluster config
func NewClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	return &Client{
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		mapper:          mapper,
		config:          config,
	}, nil
}

// GetClientset returns the underlying kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetDynamicClient returns the underlying dynamic client
func (c *Client) GetDynamicClient() *dynamic.DynamicClient {
	return c.dynamicClient
}

// GetConfig returns the rest config
func (c *Client) GetConfig() *rest.Config {
	return c.config
}
