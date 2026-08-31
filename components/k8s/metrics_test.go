package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestAggregatePodMetrics(t *testing.T) {
	pm := &metricsv1beta1.PodMetrics{
		Containers: []metricsv1beta1.ContainerMetrics{
			{
				Name: "c1",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			{
				Name: "c2",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			},
		},
	}

	cpuCores, memoryBytes := AggregatePodMetrics(pm)
	if cpuCores != 0.15 {
		t.Fatalf("cpuCores = %v, want 0.15", cpuCores)
	}
	if memoryBytes != 192*1024*1024 {
		t.Fatalf("memoryBytes = %v, want %d", memoryBytes, 192*1024*1024)
	}
}

func TestAggregatePodMetricsEmpty(t *testing.T) {
	pm := &metricsv1beta1.PodMetrics{}

	cpuCores, memoryBytes := AggregatePodMetrics(pm)
	if cpuCores != 0 || memoryBytes != 0 {
		t.Fatalf("empty pod metrics: cpuCores = %v, memoryBytes = %v, want 0/0", cpuCores, memoryBytes)
	}
}
