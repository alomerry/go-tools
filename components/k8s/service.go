package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService 创建Service
func (c *Client) CreateService(namespace string, service *corev1.Service) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
}

// UpdateService 更新Service
func (c *Client) UpdateService(namespace string, service *corev1.Service) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
}

// DeleteService 删除Service
func (c *Client) DeleteService(namespace, name string) error {
	return c.clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// GetService 获取Service
func (c *Client) GetService(namespace, name string) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
