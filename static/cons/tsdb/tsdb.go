package tsdb

type TagOp string

const (
	OpEqual         TagOp = "="
	OpNotEqual      TagOp = "!="
	OpRegexMatch    TagOp = "=~"
	OpRegexNotMatch TagOp = "!~"
)

// ToTagOp 将外部输入（如 HTTP 参数）映射为 TagOp，未识别的值回退 OpEqual
// （与既有行为一致，宽松映射由查询层做白名单校验兜底）。
func ToTagOp(op string) TagOp {
	switch op {
	case string(OpNotEqual):
		return OpNotEqual
	case string(OpRegexMatch):
		return OpRegexMatch
	case string(OpRegexNotMatch):
		return OpRegexNotMatch
	case string(OpEqual):
		return OpEqual
	default:
		return OpEqual
	}
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

// ToAggregateFn 将外部输入映射为 AggregateFn，未识别的值回退 AggMean（缺省聚合）。
func ToAggregateFn(fn string) AggregateFn {
	switch fn {
	case string(AggMax):
		return AggMax
	case string(AggMin):
		return AggMin
	case string(AggLast):
		return AggLast
	case string(AggSum):
		return AggSum
	case string(AggCount):
		return AggCount
	default:
		return AggMean
	}
}
