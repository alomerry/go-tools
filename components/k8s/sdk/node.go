package sdk

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// ListNodes lists all nodes
func (c *Client) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, opts)
}

// GetNode gets a node by name
func (c *Client) GetNode(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Node, error) {
	return c.clientset.CoreV1().Nodes().Get(ctx, name, opts)
}

// CordonNode marks the node as unschedulable
func (c *Client) CordonNode(ctx context.Context, name string) error {
	return c.setNodeSchedulable(ctx, name, false)
}

// UncordonNode marks the node as schedulable
func (c *Client) UncordonNode(ctx context.Context, name string) error {
	return c.setNodeSchedulable(ctx, name, true)
}

func (c *Client) setNodeSchedulable(ctx context.Context, name string, schedulable bool) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node, err := c.GetNode(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if node.Spec.Unschedulable == !schedulable {
			return nil // Already in desired state
		}

		node.Spec.Unschedulable = !schedulable
		_, err = c.clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return err
	})
}
