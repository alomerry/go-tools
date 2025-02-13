package k8s

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient(t *testing.T) {
	// 创建客户端
	client, err := NewClient("/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// 获取deployment
	// deployment, err := client.GetDeployment("default", "my-deployment")
	// if err != nil {
	// 	panic(err)
	// }

	// // 重启deployment
	// err = client.RestartDeployment("default", "my-deployment")
	// if err != nil {
	// 	panic(err)
	// }

	// // 获取service
	// service, err := client.GetService("default", "my-service")
	// if err != nil {
	// 	panic(err)
	// }

	// 获取pod
	pod, err := client.ListPods("kube-system", metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("pod: %+v\n", pod)
}
