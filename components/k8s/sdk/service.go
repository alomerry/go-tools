package sdk

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListServices lists services in a namespace
func (c *Client) ListServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.ServiceList, error) {
	return c.clientset.CoreV1().Services(namespace).List(ctx, opts)
}

// GetService gets a service by name
func (c *Client) GetService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, opts)
}

// CreateService creates a new service
func (c *Client) CreateService(ctx context.Context, namespace string, service *corev1.Service, opts metav1.CreateOptions) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Create(ctx, service, opts)
}

// UpdateService updates an existing service
func (c *Client) UpdateService(ctx context.Context, namespace string, service *corev1.Service, opts metav1.UpdateOptions) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Update(ctx, service, opts)
}

// DeleteService deletes a service by name
func (c *Client) DeleteService(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return c.clientset.CoreV1().Services(namespace).Delete(ctx, name, opts)
}
