package k8s

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// 样例 cAdvisor 文本（Prometheus 文本格式）覆盖：
//   - 业务容器（带 namespace/pod 标签）
//   - POD 虚拟容器（container="POD"，需排除）
//   - node/系统组件 series（无 pod 标签，需排除）
//   - 同 pod 多容器累加、科学计数法数值、带/不带时间戳
const cadvisorSample = `# HELP container_cpu_usage_seconds_total Cumulative cpu time consumed in seconds.
# TYPE container_cpu_usage_seconds_total counter
container_cpu_usage_seconds_total{beta_kubernetes_io_arch="amd64",container="POD",image="",name="k8s_POD_web-0_default_x",namespace="default",pod="web-0"} 0.42 1725000000000
container_cpu_usage_seconds_total{container="app",container_id="cri://y",image="nginx",name="k8s_app_web-0_default_y",namespace="default",pod="web-0"} 1.5 1725000000000
container_cpu_usage_seconds_total{container="sidecar",name="k8s_sidecar_web-0_default_z",namespace="default",pod="web-0"} 0.25
container_cpu_usage_seconds_total{beta_kubernetes_io_arch="amd64",container="",image="kernel",name="kernel"} 123.5
container_cpu_usage_seconds_total{container="kube-apiserver",name="apiserver",namespace="",pod=""} 99.25
container_cpu_usage_seconds_total{container="db",name="k8s_db_db-0_prod_w",namespace="prod",pod="db-0"} 2.75
# HELP container_memory_working_set_bytes Total memory in bytes.
# TYPE container_memory_working_set_bytes gauge
container_memory_working_set_bytes{container="app",name="k8s_app_web-0_default_y",namespace="default",pod="web-0"} 1.048576e+06
container_memory_working_set_bytes{container="POD",image="",name="k8s_POD_web-0_default_x",namespace="default",pod="web-0"} 4096
container_memory_working_set_bytes{container="sidecar",name="k8s_sidecar_web-0_default_z",namespace="default",pod="web-0"} 524288
container_memory_working_set_bytes{beta_kubernetes_io_arch="amd64",container="",image="kernel",name="kernel"} 2.68435456e+08
container_memory_working_set_bytes{container="db",name="k8s_db_db-0_prod_w",namespace="prod",pod="db-0"} 5.24288e+06
# HELP container_other_metric Other metric that should be ignored.
# TYPE container_other_metric gauge
container_other_metric{container="app",namespace="default",pod="web-0"} 999
`

func TestAggregateCAdvisorPodUsage(t *testing.T) {
	acc := make(map[PodRef]PodCAdvisorUsage)
	if err := aggregateCAdvisorPodUsage(strings.NewReader(cadvisorSample), acc); err != nil {
		t.Fatalf("aggregateCAdvisorPodUsage failed: %v", err)
	}

	// default/web-0：app + sidecar 两容器聚合，POD 虚拟容器排除
	//   cpu = 1.5 + 0.25 = 1.75；mem = 1048576 + 524288
	if got := acc[PodRef{Namespace: "default", Pod: "web-0"}]; got.CPUSecondsTotal != 1.75 || got.MemoryWorkingSetBytes != 1048576+524288 {
		t.Errorf("default/web-0 = %+v, want cpu 1.75 / mem %d", got, 1048576+524288)
	}

	// prod/db-0：cpu 与内存均来自单一容器 series（5.24288e+06 = 5242880）
	if got := acc[PodRef{Namespace: "prod", Pod: "db-0"}]; got.CPUSecondsTotal != 2.75 || got.MemoryWorkingSetBytes != 5242880 {
		t.Errorf("prod/db-0 = %+v, want cpu 2.75 / mem 5242880", got)
	}

	// node/系统组件 series（无 pod 标签 / container=""）不应产生任何 key
	if len(acc) != 2 {
		t.Errorf("acc length = %d, want 2 (kernel/apiserver series excluded), acc = %v", len(acc), acc)
	}
}

func TestAggregateCAdvisorPodUsageNaN(t *testing.T) {
	// NaN 样本：下游 json.Marshal 对 NaN 直接报错引发整轮失败，采集侧须跳过
	const nanSample = `# TYPE container_cpu_usage_seconds_total counter
container_cpu_usage_seconds_total{container="app",namespace="default",pod="web-0"} NaN
container_cpu_usage_seconds_total{container="sidecar",namespace="default",pod="web-0"} 0.25
# TYPE container_memory_working_set_bytes gauge
container_memory_working_set_bytes{container="app",namespace="default",pod="web-0"} NaN
`

	acc := make(map[PodRef]PodCAdvisorUsage)
	if err := aggregateCAdvisorPodUsage(strings.NewReader(nanSample), acc); err != nil {
		t.Fatalf("aggregateCAdvisorPodUsage failed: %v", err)
	}

	// 仅计非 NaN 的 sidecar：cpu=0.25，mem 缺失为 0
	if got := acc[PodRef{Namespace: "default", Pod: "web-0"}]; got.CPUSecondsTotal != 0.25 || got.MemoryWorkingSetBytes != 0 {
		t.Errorf("default/web-0 = %+v, want cpu 0.25 / mem 0 (NaN samples skipped)", got)
	}
}

func TestAggregateCAdvisorPodUsageEmptyAndInvalid(t *testing.T) {
	// 空文本：合法，得到空映射
	acc := make(map[PodRef]PodCAdvisorUsage)
	if err := aggregateCAdvisorPodUsage(strings.NewReader(""), acc); err != nil {
		t.Fatalf("empty text should parse ok: %v", err)
	}
	if len(acc) != 0 {
		t.Errorf("acc = %v, want empty", acc)
	}

	// 非 Prometheus 文本：应返回 error（fast-fail，由调用方整体重试）
	if err := aggregateCAdvisorPodUsage(strings.NewReader("<html>gateway error</html>"), acc); err == nil {
		t.Fatalf("invalid text should return error")
	}
}

// fakeNodeFetcher 构造注入 fetchPodCAdvisorUsage 的假抓取函数。
func fakeNodeFetcher(bodies map[string]string, errs map[string]error) func(context.Context, string) (io.ReadCloser, error) {
	return func(_ context.Context, node string) (io.ReadCloser, error) {
		if err := errs[node]; err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(bodies[node])), nil
	}
}

func TestFetchPodCAdvisorUsageSuccess(t *testing.T) {
	// 3 节点各报不同 pod（pod 不会跨节点），验证并发路径下的合并
	// 样例文本的 namespace 整体替换为节点专属值，使各节点产出独立 key
	bodies := map[string]string{}
	nodes := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		node := "node-" + string(rune('a'+i))
		nodes = append(nodes, node)
		body := strings.ReplaceAll(cadvisorSample, `"default"`, `"ns`+string(rune('a'+i))+`"`)
		bodies[node] = strings.ReplaceAll(body, `"prod"`, `"pr`+string(rune('a'+i))+`"`)
	}

	usage, err := fetchPodCAdvisorUsage(context.Background(), nodes, fakeNodeFetcher(bodies, nil))
	if err != nil {
		t.Fatalf("fetchPodCAdvisorUsage failed: %v", err)
	}

	// 每个节点都产出各自的 nsX/web-0，且聚合值与样例一致
	for i := 0; i < 3; i++ {
		ns := "ns" + string(rune('a'+i))
		ref := PodRef{Namespace: ns, Pod: "web-0"}
		if got := usage[ref]; got.CPUSecondsTotal != 1.75 || got.MemoryWorkingSetBytes != 1048576+524288 {
			t.Errorf("%s/web-0 = %+v, want cpu 1.75 / mem %d", ns, got, 1048576+524288)
		}
	}
	if len(usage) != 6 {
		t.Errorf("usage length = %d, want 6 (3 nodes x 2 pods)", len(usage))
	}
}

func TestFetchPodCAdvisorUsageAnyNodeFails(t *testing.T) {
	nodes := []string{"node-a", "node-b", "node-c", "node-d", "node-e", "node-f", "node-g", "node-h", "node-i", "node-j"}

	// ① 抓取失败（超 8 并发上限的节点数，任一失败整轮失败）
	bodies := map[string]string{}
	for _, n := range nodes {
		bodies[n] = cadvisorSample
	}
	errs := map[string]error{"node-i": errors.New("connection refused")}
	if _, err := fetchPodCAdvisorUsage(context.Background(), nodes, fakeNodeFetcher(bodies, errs)); err == nil {
		t.Fatalf("any-node fetch failure should fail the whole round")
	}

	// ② 解析失败（某节点返回非法文本）同样整轮失败
	badBodies := map[string]string{}
	for _, n := range nodes {
		badBodies[n] = cadvisorSample
	}
	badBodies["node-j"] = "<html>gateway error</html>"
	if _, err := fetchPodCAdvisorUsage(context.Background(), nodes, fakeNodeFetcher(badBodies, nil)); err == nil {
		t.Fatalf("any-node parse failure should fail the whole round")
	}
}

func TestMaxBytesReader(t *testing.T) {
	// 未超上限：完整读完并返回 EOF，不报错
	if _, err := io.ReadAll(&maxBytesReader{r: strings.NewReader("0123456789"), n: 10}); err != nil {
		t.Errorf("exact-limit body should read ok: %v", err)
	}

	// 超上限：返回 errCadvisorBodyTooLarge
	_, err := io.ReadAll(&maxBytesReader{r: strings.NewReader("0123456789ABC"), n: 10})
	if !errors.Is(err, errCadvisorBodyTooLarge) {
		t.Errorf("over-limit body err = %v, want %v", err, errCadvisorBodyTooLarge)
	}
}
