package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDeployment 创建Deployment
func (c *Client) CreateDeployment(namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
}

// UpdateDeployment 更新Deployment
func (c *Client) UpdateDeployment(namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
}

// DeleteDeployment 删除Deployment
func (c *Client) DeleteDeployment(namespace, name string) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// GetDeployment 获取Deployment
func (c *Client) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// RestartDeployment 重启Deployment
func (c *Client) RestartDeployment(namespace, name string) error {
	deployment, err := c.GetDeployment(namespace, name)
	if err != nil {
		return err
	}

	// 通过更新annotation的方式触发重启
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().String()

	_, err = c.UpdateDeployment(namespace, deployment)
	return err
}
