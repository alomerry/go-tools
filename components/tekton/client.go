package tekton

import (
	"fmt"
	"os"
	"path/filepath"

	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	Tekton tektonclientset.Interface
	K8s    kubernetes.Interface
}

func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fallback to default kubeconfig location
			if home := homedir.HomeDir(); home != "" {
				kubeconfig := filepath.Join(home, ".kube", "config")
				if _, statErr := os.Stat(kubeconfig); statErr == nil {
					config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create k8s config: %w", err)
	}

	tektonClient, err := tektonclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create tekton client: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &Client{
		Tekton: tektonClient,
		K8s:    k8sClient,
	}, nil
}
