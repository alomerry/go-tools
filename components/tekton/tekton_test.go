package tekton

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

// TestTektonIntegration is an integration test that requires a running K8s cluster.
// It skips if kubeconfig is not found.
func TestTektonIntegration(t *testing.T) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Skip("Skipping integration test: kubeconfig not found")
	}

	client, err := NewClient(kubeconfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	namespace := "default"

	// List Pipelines
	pipelines, err := client.ListPipelines(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		t.Logf("Failed to list pipelines (might be permission issue or CRD missing): %v", err)
	} else {
		t.Logf("Found %d pipelines", len(pipelines.Items))
	}

	// List PipelineRuns
	prs, err := client.ListPipelineRuns(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		t.Logf("Failed to list pipelineruns: %v", err)
	} else {
		t.Logf("Found %d pipelineruns", len(prs.Items))
		for _, pr := range prs.Items {
			status := client.GetPipelineRunStatus(&pr)
			t.Logf("PipelineRun %s: %s", pr.Name, status)
			
			// Try to get logs for the first one
			if len(prs.Items) > 0 {
				logs, err := client.GetPipelineRunLogs(ctx, namespace, pr.Name)
				if err != nil {
					t.Logf("Failed to get logs: %v", err)
				} else {
					t.Logf("Got logs for %d taskruns", len(logs))
				}
			}
		}
	}
}
