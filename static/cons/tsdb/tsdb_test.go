package tsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTagOp(t *testing.T) {
	assert.Equal(t, OpEqual, ToTagOp("="))
	assert.Equal(t, OpNotEqual, ToTagOp("!="))
	assert.Equal(t, OpRegexMatch, ToTagOp("=~"))
	assert.Equal(t, OpRegexNotMatch, ToTagOp("!~"))
	// 未知值与空串回退 OpEqual，保持既有宽松映射行为
	assert.Equal(t, OpEqual, ToTagOp("unknown"))
	assert.Equal(t, OpEqual, ToTagOp(""))
}

func TestToAggregateFn(t *testing.T) {
	assert.Equal(t, AggMean, ToAggregateFn("mean"))
	assert.Equal(t, AggMax, ToAggregateFn("max"))
	assert.Equal(t, AggMin, ToAggregateFn("min"))
	assert.Equal(t, AggLast, ToAggregateFn("last"))
	assert.Equal(t, AggSum, ToAggregateFn("sum"))
	assert.Equal(t, AggCount, ToAggregateFn("count"))
	// 未知值与空串回退缺省聚合 mean
	assert.Equal(t, AggMean, ToAggregateFn("unknown"))
	assert.Equal(t, AggMean, ToAggregateFn(""))
}

func TestIsValidTagOp(t *testing.T) {
	assert.True(t, IsValidTagOp("="))
	assert.True(t, IsValidTagOp("!="))
	assert.True(t, IsValidTagOp("=~"))
	assert.True(t, IsValidTagOp("!~"))
	assert.False(t, IsValidTagOp(""))
	assert.False(t, IsValidTagOp("~"))
	assert.False(t, IsValidTagOp("=="))
}

// TestToEmptyFill 空值窗口三态映射：已知值原样返回（字符串形态与 homelab-backend
// proto 枚举值名 EMPTY_FILL_* 逐字对齐）；未识别值与空串回退零值 ""（未指定语义，
// 查询层缺省 CONNECT，不做非法值兜底改写——显式未知值由查询层校验拒绝）。
func TestToEmptyFill(t *testing.T) {
	assert.Equal(t, EmptyFillConnect, ToEmptyFill("EMPTY_FILL_CONNECT"))
	assert.Equal(t, EmptyFillZero, ToEmptyFill("EMPTY_FILL_ZERO"))
	assert.Equal(t, EmptyFillGap, ToEmptyFill("EMPTY_FILL_GAP"))
	// 空串/未识别（含 proto 的 UNSPECIFIED 名，由调用方归一为空串后不会到达；
	// 若直传也按未识别回退未指定语义）
	assert.Equal(t, EmptyFill(""), ToEmptyFill(""))
	assert.Equal(t, EmptyFill(""), ToEmptyFill("EMPTY_FILL_UNSPECIFIED"))
	assert.Equal(t, EmptyFill(""), ToEmptyFill("zero"))
	assert.Equal(t, EmptyFill(""), ToEmptyFill("connect "))
}

// TestIsValidEmptyFill 空值窗口三态白名单：空串表示未指定，不在白名单内。
func TestIsValidEmptyFill(t *testing.T) {
	assert.True(t, IsValidEmptyFill("EMPTY_FILL_CONNECT"))
	assert.True(t, IsValidEmptyFill("EMPTY_FILL_ZERO"))
	assert.True(t, IsValidEmptyFill("EMPTY_FILL_GAP"))
	assert.False(t, IsValidEmptyFill(""))
	assert.False(t, IsValidEmptyFill("EMPTY_FILL_UNSPECIFIED"))
	assert.False(t, IsValidEmptyFill("zero"))
	assert.False(t, IsValidEmptyFill(`gap) |> fill(column: "_value", value: 1.0`))
}

func TestIsValidAggregateFn(t *testing.T) {
	assert.True(t, IsValidAggregateFn("mean"))
	assert.True(t, IsValidAggregateFn("max"))
	assert.True(t, IsValidAggregateFn("min"))
	assert.True(t, IsValidAggregateFn("last"))
	assert.True(t, IsValidAggregateFn("sum"))
	assert.True(t, IsValidAggregateFn("count"))
	assert.False(t, IsValidAggregateFn(""))
	assert.False(t, IsValidAggregateFn("avg"))
	assert.False(t, IsValidAggregateFn(`mean) or (x`))
}
