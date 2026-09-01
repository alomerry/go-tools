package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cAdvisor 关注的指标名（Prometheus 文本格式）：
// CPU 为累计计数器（秒），由查询侧按 non_negative_derivative 计算速率；
// 内存为瞬时值（字节）。
const (
	cadvisorCPUMetric    = "container_cpu_usage_seconds_total"
	cadvisorMemoryMetric = "container_memory_working_set_bytes"
)

// cAdvisor series 的关键标签
const (
	cadvisorLabelContainer = "container"
	cadvisorLabelNamespace = "namespace"
	cadvisorLabelPod       = "pod"
)

// cadvisorPODContainer k8s 为每个 pod 注入的 pause/infra 虚拟容器名（内含所有容器份额之和），
// 无真实业务负载，聚合时排除，避免重复计算。
const cadvisorPODContainer = "POD"

const (
	// cadvisorFetchConcurrency 单轮抓取节点的并发上限（errgroup.SetLimit）
	cadvisorFetchConcurrency = 8

	// cadvisorMaxBodyBytes 单节点响应体大小上限（64MB）。
	// 正常 cAdvisor 输出为数百 KB 量级，超限视为异常（如代理返回非预期内容），立即报错避免内存放大。
	cadvisorMaxBodyBytes = int64(64) << 20
)

// errCadvisorBodyTooLarge 节点响应体超过 cadvisorMaxBodyBytes 上限
var errCadvisorBodyTooLarge = errors.New("cadvisor metrics body too large")

// PodRef pod 引用（namespace + pod 名），作为 cAdvisor 用量映射的 key。
type PodRef struct {
	Namespace string
	Pod       string
}

// PodCAdvisorUsage 单个 pod 的 cAdvisor 用量（业务容器聚合，不含 POD 虚拟容器）。
type PodCAdvisorUsage struct {
	// CPUSecondsTotal container_cpu_usage_seconds_total 累计计数器之和（秒）。
	// pod 重启归零（回绕）不在采集侧处理，由查询侧 non_negative_derivative 兜底。
	CPUSecondsTotal float64
	// MemoryWorkingSetBytes container_memory_working_set_bytes 之和（字节）
	MemoryWorkingSetBytes int64
}

// GetPodCAdvisorUsage 通过 apiserver proxy 逐节点拉取 kubelet cAdvisor 指标
// （GET /api/v1/nodes/{node}/proxy/metrics/cadvisor，Prometheus 文本格式），
// 按 pod 维度（namespace/pod）聚合业务容器用量，返回 PodRef -> 用量映射。
// 仅聚合带 namespace+pod 标签且 container != "POD" 的业务容器 series，
// 过滤掉无 pod 标签的 node/系统组件 series 与 POD 虚拟容器。
// 任一节点抓取/解析失败即取消其余在途请求并返回 error（fast-fail），由调用方整体重试，不做部分写入。
func (c *Client) GetPodCAdvisorUsage(ctx context.Context) (map[PodRef]PodCAdvisorUsage, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes failed: %w", err)
	}

	names := make([]string, len(nodes.Items))
	for i := range nodes.Items {
		names[i] = nodes.Items[i].Name
	}
	return fetchPodCAdvisorUsage(ctx, names, c.proxyNodeCAdvisorMetrics)
}

// fetchPodCAdvisorUsage 有界并发抓取各节点 cAdvisor 指标并按 pod 聚合。
// fetch 为单节点抓取函数（参数化便于注入假实现做单测）；任一节点失败即经 errgroup 取消其余
// 在途请求并整轮失败（总超时由调用方 ctx 控制）；各 goroutine 聚合进独立 map，Wait 后单协程合并，无共享写。
func fetchPodCAdvisorUsage(ctx context.Context, nodes []string, fetch func(ctx context.Context, node string) (io.ReadCloser, error)) (map[PodRef]PodCAdvisorUsage, error) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(cadvisorFetchConcurrency)

	locals := make([]map[PodRef]PodCAdvisorUsage, len(nodes))
	for i, node := range nodes {
		g.Go(func() error {
			body, err := fetch(gctx, node)
			if err != nil {
				return fmt.Errorf("proxy cadvisor metrics of node %s failed: %w", node, err)
			}
			defer body.Close()

			local := make(map[PodRef]PodCAdvisorUsage)
			if err := aggregateCAdvisorPodUsage(&maxBytesReader{r: body, n: cadvisorMaxBodyBytes}, local); err != nil {
				return fmt.Errorf("parse cadvisor metrics of node %s failed: %w", node, err)
			}
			locals[i] = local
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// pod 不会跨节点，合并实际为写入；仍按求和防御性处理
	usage := make(map[PodRef]PodCAdvisorUsage)
	for _, local := range locals {
		for ref, u := range local {
			merged := usage[ref]
			merged.CPUSecondsTotal += u.CPUSecondsTotal
			merged.MemoryWorkingSetBytes += u.MemoryWorkingSetBytes
			usage[ref] = merged
		}
	}
	return usage, nil
}

// proxyNodeCAdvisorMetrics 通过 apiserver proxy 请求节点的 cAdvisor 指标接口，
// 复用 Client 的 rest.Config（SA token/TLS），不直连 kubelet。
// 使用 Stream 流式返回响应体（非 2xx 已由 Stream 转为 error），避免整体 ReadAll 的双份内存。
// 超时由 ctx 控制。
//
// 防回归：AbsPath 不做路径转义，node 必须来自 Nodes().List（apiserver 已校验 DNS-1123 合法性），
// 禁止传入任何外部输入，否则可能被拼接出任意 proxy 路径。
func (c *Client) proxyNodeCAdvisorMetrics(ctx context.Context, node string) (io.ReadCloser, error) {
	return c.clientset.CoreV1().RESTClient().
		Get().
		AbsPath("/api/v1/nodes/" + node + "/proxy/metrics/cadvisor").
		Stream(ctx)
}

// aggregateCAdvisorPodUsage 解析 Prometheus 文本格式的 cAdvisor 指标，
// 将关注指标的 pod 维度用量累加进 acc（跨调用累加同 key 直接求和）。
// cAdvisor 单节点输出数百个指标、数千行，expfmt 按 MetricFamily 解码（v0.67.5 内部为全量解析，
// 流式只体现在 HTTP 读取层，上限由 maxBytesReader 兜底），
// 仅对关注指标做标签提取与累加，其余 family 直接跳过。
func aggregateCAdvisorPodUsage(r io.Reader, acc map[PodRef]PodCAdvisorUsage) error {
	dec := expfmt.NewDecoder(r, expfmt.FmtText)
	mf := &dto.MetricFamily{}
	for {
		if err := dec.Decode(mf); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("decode prometheus text failed: %w", err)
		}

		name := mf.GetName()
		if name != cadvisorCPUMetric && name != cadvisorMemoryMetric {
			continue
		}

		for _, m := range mf.GetMetric() {
			labels := metricLabels(m)
			container := labels[cadvisorLabelContainer]
			namespace := labels[cadvisorLabelNamespace]
			pod := labels[cadvisorLabelPod]
			// 排除 POD 虚拟容器；无 namespace/pod 标签的 series 为 node/系统组件维度，非业务容器
			if container == "" || container == cadvisorPODContainer || namespace == "" || pod == "" {
				continue
			}

			// NaN 会让下游 json.Marshal 直接报错引发整轮失败，采集侧跳过该样本
			ref := PodRef{Namespace: namespace, Pod: pod}
			u := acc[ref]
			switch name {
			case cadvisorCPUMetric:
				if v := m.GetCounter().GetValue(); !math.IsNaN(v) {
					u.CPUSecondsTotal += v
				}
			case cadvisorMemoryMetric:
				if v := m.GetGauge().GetValue(); !math.IsNaN(v) {
					u.MemoryWorkingSetBytes += int64(v)
				}
			}
			acc[ref] = u
		}
	}
}

// maxBytesReader 读取超过上限即返回 errCadvisorBodyTooLarge 的 io.Reader
// （io.LimitReader 到限会静默 EOF，无法表达「超限报错」语义）。
// 超限在累计读取量超过 n 时触发，单次读请求可能超出上限一个 chunk，量级可忽略。
type maxBytesReader struct {
	r io.Reader
	n int64
}

func (m *maxBytesReader) Read(p []byte) (int, error) {
	got, err := m.r.Read(p)
	m.n -= int64(got)
	if m.n < 0 {
		return got, errCadvisorBodyTooLarge
	}
	return got, err
}

// metricLabels 提取单个 sample 的标签集
func metricLabels(m *dto.Metric) map[string]string {
	labels := make(map[string]string, len(m.GetLabel()))
	for _, lp := range m.GetLabel() {
		labels[lp.GetName()] = lp.GetValue()
	}
	return labels
}
