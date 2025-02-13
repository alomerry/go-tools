package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPod 获取Pod
func (c *Client) GetPod(namespace, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// DeletePod 删除Pod
func (c *Client) DeletePod(namespace, name string) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// ListPods 列出Pod
func (c *Client) ListPods(namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(context.TODO(), opts)
}

// RestartPod 重启Pod (通过删除Pod的方式)
func (c *Client) RestartPod(namespace, name string) error {
	return c.DeletePod(namespace, name)
}
