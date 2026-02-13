package tekton

import (
	"context"
	"fmt"
	"sort"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// ListPipelineRuns returns a list of PipelineRuns in the specified namespace.
func (c *Client) ListPipelineRuns(ctx context.Context, namespace string, opts metav1.ListOptions) (*pipelinev1.PipelineRunList, error) {
	return c.Tekton.TektonV1().PipelineRuns(namespace).List(ctx, opts)
}

// GetPipelineRun returns a specific PipelineRun.
func (c *Client) GetPipelineRun(ctx context.Context, namespace, name string) (*pipelinev1.PipelineRun, error) {
	return c.Tekton.TektonV1().PipelineRuns(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreatePipelineRun creates a new PipelineRun.
func (c *Client) CreatePipelineRun(ctx context.Context, namespace string, pr *pipelinev1.PipelineRun) (*pipelinev1.PipelineRun, error) {
	return c.Tekton.TektonV1().PipelineRuns(namespace).Create(ctx, pr, metav1.CreateOptions{})
}

// CancelPipelineRun cancels a running PipelineRun.
func (c *Client) CancelPipelineRun(ctx context.Context, namespace, name string) error {
	pr, err := c.GetPipelineRun(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update status to Cancelled
	// In v1, we set Spec.Status to "Cancelled" to cancel the run
	pr.Spec.Status = pipelinev1.PipelineRunSpecStatusCancelled
	_, err = c.Tekton.TektonV1().PipelineRuns(namespace).Update(ctx, pr, metav1.UpdateOptions{})
	return err
}

// DeletePipelineRun deletes a PipelineRun.
func (c *Client) DeletePipelineRun(ctx context.Context, namespace, name string) error {
	return c.Tekton.TektonV1().PipelineRuns(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// SimpleStatus represents a simplified status of a PipelineRun.
type SimpleStatus string

const (
	StatusRunning   SimpleStatus = "Running"
	StatusSucceeded SimpleStatus = "Succeeded"
	StatusFailed    SimpleStatus = "Failed"
	StatusCancelled SimpleStatus = "Cancelled"
	StatusUnknown   SimpleStatus = "Unknown"
)

// GetPipelineRunStatus returns the simplified status of a PipelineRun.
func (c *Client) GetPipelineRunStatus(pr *pipelinev1.PipelineRun) SimpleStatus {
	if pr == nil {
		return StatusUnknown
	}
	
	s := pr.Status.GetCondition(apis.ConditionSucceeded)
	if s == nil {
		return StatusUnknown
	}

	switch s.Status {
	case corev1.ConditionUnknown:
		if pr.Spec.Status == pipelinev1.PipelineRunSpecStatusCancelled {
			return StatusCancelled
		}
		return StatusRunning
	case corev1.ConditionTrue:
		return StatusSucceeded
	case corev1.ConditionFalse:
		// Check if it was cancelled
		if pr.Spec.Status == pipelinev1.PipelineRunSpecStatusCancelled {
			return StatusCancelled
		}
		// Also check the reason in condition
		if s.Reason == "PipelineRunCancelled" {
			return StatusCancelled
		}
		return StatusFailed
	default:
		return StatusUnknown
	}
}

// GetLatestPipelineRun returns the most recently started PipelineRun for a given pipeline name.
func (c *Client) GetLatestPipelineRun(ctx context.Context, namespace, pipelineName string) (*pipelinev1.PipelineRun, error) {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipeline=%s", pipelineName),
	}
	prList, err := c.ListPipelineRuns(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}

	if len(prList.Items) == 0 {
		return nil, fmt.Errorf("no pipelineruns found for pipeline %s", pipelineName)
	}

	// Sort by StartTime descending
	sort.Slice(prList.Items, func(i, j int) bool {
		t1 := prList.Items[i].Status.StartTime
		t2 := prList.Items[j].Status.StartTime
		if t1 == nil {
			return false
		}
		if t2 == nil {
			return true
		}
		return t1.Time.After(t2.Time)
	})

	return &prList.Items[0], nil
}
