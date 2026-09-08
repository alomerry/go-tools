package cat

import (
	"unicode/utf8"

	"github.com/alomerry/go-tools/components/tsdb"
)

// 点位字段长度上限，防止异常输入（大响应体、长堆栈等）撑爆内存。
const (
	maxProblemMessage  = 512
	maxEventData       = 2048
	maxTransactionData = 8
	maxDataValue       = 1024
)

// point 是 cat 组件产出的点位，统一经 tsdb Metric → Kafka 异步落 InfluxDB。
type point struct {
	measurement string
	tags        map[string]string
	fields      map[string]any
}

// options 将点位转换为 tsdb 的 functional options。tag 走 WithTag；
// field 统一走 WithFieldAny 直写 Fields（含字符串类型，绕过 string→Tags
// 的隐式映射）。
func (p point) options() []func(any) {
	opts := make([]func(any), 0, len(p.tags)+len(p.fields)+1)
	opts = append(opts, tsdb.WithMetric(p.measurement))
	for k, v := range p.tags {
		opts = append(opts, tsdb.WithTag(k, v))
	}
	for k, v := range p.fields {
		opts = append(opts, tsdb.WithFieldAny(k, v))
	}
	return opts
}

// emit 点位出口。包级变量仅为单测注入捕获点位，生产行为固定走 tsdb 异步链路
// （LogForCnt 内补 cnt=1 与时间戳；writer 未就绪或池满时由 tsdb 非阻塞丢弃）。
var emit = func(p point) {
	tsdb.NewMetric(p.options()...).LogForCnt()
}

// truncateString 按字节数截断，回退到 UTF-8 字符边界，不切断多字节字符。
// 与 utils/string.Limit（字节直切，可能切断多字节字符）语义不同，故不合并：
// 点位 field 须经 InfluxDB/Grafana 展示，须保证合法 UTF-8。
func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max]
}
