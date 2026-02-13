package sdk

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	return c.clientset.CoreV1().Namespaces().List(ctx, opts)
}

// ListPersistentVolumes lists all persistent volumes
func (c *Client) ListPersistentVolumes(ctx context.Context, opts metav1.ListOptions) (*corev1.PersistentVolumeList, error) {
	return c.clientset.CoreV1().PersistentVolumes().List(ctx, opts)
}

// ListPersistentVolumeClaims lists persistent volume claims in a namespace
func (c *Client) ListPersistentVolumeClaims(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	return c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
}

// ListConfigMaps lists config maps in a namespace
func (c *Client) ListConfigMaps(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.ConfigMapList, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, opts)
}

// ListSecrets lists secrets in a namespace
func (c *Client) ListSecrets(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.SecretList, error) {
	return c.clientset.CoreV1().Secrets(namespace).List(ctx, opts)
}
