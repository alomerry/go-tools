package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// GetPodMetrics 获取命名空间下所有 pod 的指标（依赖集群已安装 metrics-server）
// namespace 传空串表示所有命名空间
func (c *Client) GetPodMetrics(ctx context.Context, namespace string) (*metricsv1beta1.PodMetricsList, error) {
	return c.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
}

// AggregatePodMetrics 聚合单个 pod 的 CPU/内存用量
// 返回 cpuCores（核数，float64）与 memoryBytes（字节数，int64）
func AggregatePodMetrics(pm *metricsv1beta1.PodMetrics) (float64, int64) {
	var cpuMilliCores int64
	var memoryBytes int64
	for i := range pm.Containers {
		cpuMilliCores += pm.Containers[i].Usage.Cpu().MilliValue()
		memoryBytes += pm.Containers[i].Usage.Memory().Value()
	}
	return float64(cpuMilliCores) / 1000, memoryBytes
}
