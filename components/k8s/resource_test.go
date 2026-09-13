package k8s

import (
	"testing"

	"github.com/alomerry/go-tools/test"
)

// TestCreateResourceByYaml 依赖真实集群（kubeconfig + apiserver 可达），集成开关控制
func TestCreateResourceByYaml(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	client, err := NewClient("xxx")
	if err != nil {
		t.Fatal(err)
	}

	yamlContent := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
`

	client.CreateResourceByYaml(yamlContent)
}
