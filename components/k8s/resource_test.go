package k8s

import (
	"testing"
)

func TestCreateResourceByYaml(t *testing.T) {
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
