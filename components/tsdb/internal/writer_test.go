package internal

import (
	"testing"
	"time"
)

type testSerializer struct{}

func (testSerializer) Encode() ([]byte, error) { return []byte("{}"), nil }
func (testSerializer) Decode([]byte) error     { return nil }

// TestAsyncWriteUninitializedNotBlocking writer 未初始化（pool 为 nil）时，
// AsyncWrite 必须立即走丢弃分支，不得阻塞（向 nil channel 发送原本会永久阻塞）。
func TestAsyncWriteUninitializedNotBlocking(t *testing.T) {
	pool = nil

	done := make(chan struct{})
	go func() {
		AsyncWrite(testSerializer{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("AsyncWrite on uninitialized pool blocked")
	}
}

// TestAsyncWritePoolFullDropsAndNotBlocking 池满时丢弃新点位并立即返回，不阻塞。
func TestAsyncWritePoolFullDropsAndNotBlocking(t *testing.T) {
	pool = make(chan Serializer, 1)
	defer func() { pool = nil }()

	AsyncWrite(testSerializer{}) // 占满
	AsyncWrite(testSerializer{}) // 池满：丢弃并立即返回

	if len(pool) != 1 {
		t.Fatalf("pool len = %d, want 1 (full pool must not accept more)", len(pool))
	}
}

// TestWarnDropRateLimited 丢弃告警按 dropWarnInterval 限频。
func TestWarnDropRateLimited(t *testing.T) {
	dropMu.Lock()
	lastDrop = time.Time{}
	dropMu.Unlock()

	warnDrop()
	dropMu.Lock()
	first := lastDrop
	dropMu.Unlock()

	if first.IsZero() {
		t.Fatal("first warnDrop should set lastDrop")
	}

	warnDrop() // 间隔内不应更新
	dropMu.Lock()
	second := lastDrop
	dropMu.Unlock()

	if !first.Equal(second) {
		t.Fatal("warnDrop within interval should be rate limited")
	}
}
