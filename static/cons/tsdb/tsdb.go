package tsdb

type TagOp string

const (
	OpEqual         TagOp = "="
	OpNotEqual      TagOp = "!="
	OpRegexMatch    TagOp = "=~"
	OpRegexNotMatch TagOp = "!~"
)

// IsValidTagOp 判定 op 是否为已知运算符（白名单），供外部输入校验复用。
func IsValidTagOp(op string) bool {
	switch TagOp(op) {
	case OpEqual, OpNotEqual, OpRegexMatch, OpRegexNotMatch:
		return true
	default:
		return false
	}
}

// ToTagOp 将外部输入（如 HTTP 参数）映射为 TagOp：已知值原样返回，未识别的值
// 回退 OpEqual（与既有行为一致，宽松映射由查询层白名单校验兜底）。
func ToTagOp(op string) TagOp {
	if IsValidTagOp(op) {
		return TagOp(op)
	}
	return OpEqual
}

// AggregateFn 是 Flux aggregateWindow 的聚合函数名。
type AggregateFn string

const (
	AggMean  AggregateFn = "mean"
	AggMax   AggregateFn = "max"
	AggMin   AggregateFn = "min"
	AggLast  AggregateFn = "last"
	AggSum   AggregateFn = "sum"
	AggCount AggregateFn = "count"
)

// IsValidAggregateFn 判定 fn 是否为已知聚合函数（白名单），供外部输入校验复用。
func IsValidAggregateFn(fn string) bool {
	switch AggregateFn(fn) {
	case AggMean, AggMax, AggMin, AggLast, AggSum, AggCount:
		return true
	default:
		return false
	}
}

// ToAggregateFn 将外部输入映射为 AggregateFn：已知值原样返回，未识别的值
// 回退 AggMean（缺省聚合）。
func ToAggregateFn(fn string) AggregateFn {
	if IsValidAggregateFn(fn) {
		return AggregateFn(fn)
	}
	return AggMean
}

// EmptyFill aggregateWindow 空值窗口的处理策略（monitor-opt2 曲线空值三态，
// 类 Grafana fill）。常量字符串与 homelab-backend proto MetricConfig.empty_fill
// 的枚举值名（EMPTY_FILL_*）逐字对齐，作为跨仓库传递的唯一字符串形态。
type EmptyFill string

const (
	// EmptyFillConnect 连接（缺省语义）：createEmpty: false——空窗口不产出行，
	// 前端把相邻点直接连线；与历史查询产物逐字节一致（零回归基线）。
	EmptyFillConnect EmptyFill = "EMPTY_FILL_CONNECT"
	// EmptyFillZero 补零：createEmpty: true 且追加 fill(column: "_value", value: 0.0)，
	// 空窗口以 0 填充（count 类聚合无数据=0 语义正确）。
	EmptyFillZero EmptyFill = "EMPTY_FILL_ZERO"
	// EmptyFillGap 断开留空：createEmpty: true 不加 fill，空窗口 _value 为
	// null，行解析（nil）与 protojson 序列化（NullValue→JSON null）逐层透传，
	// 前端以 null 断线渲染。
	EmptyFillGap EmptyFill = "EMPTY_FILL_GAP"
)

// IsValidEmptyFill 判定 s 是否为已知空值处理策略（白名单）；空串表示未指定，
// 不在白名单内（由查询层按缺省 CONNECT 兜底）。
func IsValidEmptyFill(s string) bool {
	switch EmptyFill(s) {
	case EmptyFillConnect, EmptyFillZero, EmptyFillGap:
		return true
	default:
		return false
	}
}

// ToEmptyFill 将外部输入映射为 EmptyFill：已知值原样返回；未识别值与空串
// 回退零值 ""（未指定语义，查询层缺省按 CONNECT 生成，与历史行为一致）。
// 注意与 ToTagOp/ToAggregateFn 的差异：本类型没有"非法值兜底成某个合法值"
// 的强需求——未指定与 CONNECT 查询产物相同，回退零值即可；显式传入未知值
// 由 GetQuery/Query 出口的 validateEmptyFill 拒绝，避免笔误静默变成缺省。
func ToEmptyFill(s string) EmptyFill {
	if IsValidEmptyFill(s) {
		return EmptyFill(s)
	}
	return ""
}
