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
