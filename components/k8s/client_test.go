package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alomerry/go-tools/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestClient 依赖真实集群（kubeconfig + apiserver 可达），集成开关控制
func TestClient(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	if _, err := os.Stat(kubeconfig); err != nil {
		t.Skipf("kubeconfig %q not available: %v", kubeconfig, err)
	}

	// 创建客户端
	client, err := NewClient(kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	// 获取pod
	pod, err := client.ListPods("kube-system", metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("pod: %+v\n", pod)
}
