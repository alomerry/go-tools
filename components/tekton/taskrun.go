package tekton

import (
	"context"
	"fmt"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListTaskRuns returns a list of TaskRuns.
func (c *Client) ListTaskRuns(ctx context.Context, namespace string, opts metav1.ListOptions) (*pipelinev1.TaskRunList, error) {
	return c.Tekton.TektonV1().TaskRuns(namespace).List(ctx, opts)
}

// ListTaskRunsForPipelineRun returns all TaskRuns for a specific PipelineRun.
func (c *Client) ListTaskRunsForPipelineRun(ctx context.Context, namespace, pipelineRunName string) (*pipelinev1.TaskRunList, error) {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", pipelineRunName),
	}
	return c.ListTaskRuns(ctx, namespace, opts)
}

// GetTaskRun returns a specific TaskRun.
func (c *Client) GetTaskRun(ctx context.Context, namespace, name string) (*pipelinev1.TaskRun, error) {
	return c.Tekton.TektonV1().TaskRuns(namespace).Get(ctx, name, metav1.GetOptions{})
}
