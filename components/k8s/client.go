package k8s

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

func NewConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		// 使用提供的kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// 在集群内部运行时使用集群内配置
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}
	return config, nil
}

// NewClient 创建一个新的 K8s客户端
func NewClient(kubeconfig string) (*Client, error) {
	config, err := NewConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 初始化动态客户端
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 初始化发现客户端
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	// 初始化 RESTMapper
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	return &Client{
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		mapper:          mapper,
		config:          config,
	}, nil
}

func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

func (c *Client) GetDynamicClient() *dynamic.DynamicClient {
	return c.dynamicClient
}

func (c *Client) GetDiscoveryClient() *discovery.DiscoveryClient {
	return c.discoveryClient
}

func (c *Client) GetMapper() *restmapper.DeferredDiscoveryRESTMapper {
	return c.mapper
}

func (c *Client) GetConfig() *rest.Config {
	return c.config
}
