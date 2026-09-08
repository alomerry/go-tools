package tsdb

import (
	"testing"
)

// TestWithFieldAnyWritesFields 验证 WithFieldAny 直写 Fields（含 string 类型），
// 不走 string→Tags 的隐式映射；对比 WithField("k", string) 仍进 Tags。
func TestWithFieldAnyWritesFields(t *testing.T) {
	m := newMetric(
		WithMetric("event"),
		WithField("status", "ok"),       // string 经 WithField → Tags（既有隐式行为）
		WithFieldAny("data", "payload"), // string 经 WithFieldAny → Fields
		WithFieldAny("num", int64(3)),   // 数值经 WithFieldAny → Fields
	)

	if m.Tags["status"] != "ok" {
		t.Fatalf("Tags[status] = %q, want ok (implicit string->tag kept)", m.Tags["status"])
	}
	if _, ok := m.Tags["data"]; ok {
		t.Fatal("WithFieldAny string must not go to Tags")
	}
	if m.Fields["data"] != "payload" {
		t.Fatalf("Fields[data] = %v, want payload", m.Fields["data"])
	}
	if m.Fields["num"] != int64(3) {
		t.Fatalf("Fields[num] = %v, want 3", m.Fields["num"])
	}
}

// TestWithFieldAnySkipsEmptyAndNil 空 key 与 nil 值跳过，不 panic。
func TestWithFieldAnySkipsEmptyAndNil(t *testing.T) {
	m := newMetric(
		WithFieldAny("", "v"),
		WithFieldAny("k", nil),
	)
	if len(m.Fields) != 0 {
		t.Fatalf("Fields = %+v, want empty", m.Fields)
	}
}
