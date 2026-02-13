package tekton

import (
	"context"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPipelines returns a list of Pipelines.
func (c *Client) ListPipelines(ctx context.Context, namespace string, opts metav1.ListOptions) (*pipelinev1.PipelineList, error) {
	return c.Tekton.TektonV1().Pipelines(namespace).List(ctx, opts)
}

// GetPipeline returns a specific Pipeline.
func (c *Client) GetPipeline(ctx context.Context, namespace, name string) (*pipelinev1.Pipeline, error) {
	return c.Tekton.TektonV1().Pipelines(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreatePipeline creates a new Pipeline.
func (c *Client) CreatePipeline(ctx context.Context, namespace string, p *pipelinev1.Pipeline) (*pipelinev1.Pipeline, error) {
	return c.Tekton.TektonV1().Pipelines(namespace).Create(ctx, p, metav1.CreateOptions{})
}

// UpdatePipeline updates an existing Pipeline.
func (c *Client) UpdatePipeline(ctx context.Context, namespace string, p *pipelinev1.Pipeline) (*pipelinev1.Pipeline, error) {
	return c.Tekton.TektonV1().Pipelines(namespace).Update(ctx, p, metav1.UpdateOptions{})
}

// DeletePipeline deletes a Pipeline.
func (c *Client) DeletePipeline(ctx context.Context, namespace, name string) error {
	return c.Tekton.TektonV1().Pipelines(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ListTasks returns a list of Tasks.
func (c *Client) ListTasks(ctx context.Context, namespace string, opts metav1.ListOptions) (*pipelinev1.TaskList, error) {
	return c.Tekton.TektonV1().Tasks(namespace).List(ctx, opts)
}

// GetTask returns a specific Task.
func (c *Client) GetTask(ctx context.Context, namespace, name string) (*pipelinev1.Task, error) {
	return c.Tekton.TektonV1().Tasks(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateTask creates a new Task.
func (c *Client) CreateTask(ctx context.Context, namespace string, t *pipelinev1.Task) (*pipelinev1.Task, error) {
	return c.Tekton.TektonV1().Tasks(namespace).Create(ctx, t, metav1.CreateOptions{})
}

// UpdateTask updates an existing Task.
func (c *Client) UpdateTask(ctx context.Context, namespace string, t *pipelinev1.Task) (*pipelinev1.Task, error) {
	return c.Tekton.TektonV1().Tasks(namespace).Update(ctx, t, metav1.UpdateOptions{})
}

// DeleteTask deletes a Task.
func (c *Client) DeleteTask(ctx context.Context, namespace, name string) error {
	return c.Tekton.TektonV1().Tasks(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
