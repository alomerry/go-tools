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
