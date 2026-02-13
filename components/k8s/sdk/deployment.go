package sdk

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// ListDeployments lists deployments in a namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string, opts metav1.ListOptions) (*appsv1.DeploymentList, error) {
	return c.clientset.AppsV1().Deployments(namespace).List(ctx, opts)
}

// GetDeployment gets a deployment by name
func (c *Client) GetDeployment(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, opts)
}

// CreateDeployment creates a new deployment
func (c *Client) CreateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment, opts metav1.CreateOptions) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, opts)
}

// UpdateDeployment updates an existing deployment
func (c *Client) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment, opts metav1.UpdateOptions) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, opts)
}

// DeleteDeployment deletes a deployment by name
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, opts)
}

// ScaleDeployment scales a deployment to a specific number of replicas
func (c *Client) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) (*appsv1.Deployment, error) {
	scale, err := c.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return c.GetDeployment(ctx, namespace, name, metav1.GetOptions{})
}

// RestartDeployment restarts a deployment by updating its annotation
func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := c.GetDeployment(ctx, namespace, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		_, err = c.UpdateDeployment(ctx, namespace, deployment, metav1.UpdateOptions{})
		return err
	})
}

// RollbackDeployment rolls back a deployment to a specific revision
// Note: This is a simplified implementation. True rollback logic in kubectl is complex.
// For now, we might need to rely on the user providing the full previous spec or implement specific logic later.
// A placeholder is provided here if we want to implement it using `deployment.Spec.RollbackTo` which is deprecated/removed in AppsV1.
// The correct way is to find the ReplicaSet for the revision and apply its template to the Deployment.
func (c *Client) RollbackDeployment(ctx context.Context, namespace, name string, revision int64) error {
	// TODO: Implement full rollback logic similar to `kubectl rollout undo`
	return fmt.Errorf("rollback not implemented for apps/v1 deployments in this SDK yet")
}
