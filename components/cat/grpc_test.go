package cat

import (
	"context"
	"errors"
	"testing"

	"github.com/alomerry/go-tools/static/cons"
	"google.golang.org/grpc"
)

func unaryInfo() *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: "/backend.svc.Svc/Method"}
}

func TestUnaryServerInterceptorSuccess(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	iv := UnaryServerInterceptor(nil)
	resp, err := iv(context.Background(), "req", unaryInfo(), func(ctx context.Context, req any) (any, error) {
		if ctx.Value(cons.TraceIdKey) == "" {
			t.Error("trace id missing in handler ctx")
		}
		if TraceIdFromCtx(ctx) == "" {
			t.Error("TraceIdFromCtx empty in handler ctx")
		}
		if req != "req" {
			t.Errorf("req = %v, want req", req)
		}
		return "resp", nil
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if resp != "resp" {
		t.Fatalf("resp = %v, want resp", resp)
	}

	p := drainPoint(t, ch)
	assertNoPoint(t, ch)

	if p.measurement != "transaction" {
		t.Fatalf("measurement = %q, want transaction", p.measurement)
	}
	if p.tags["service"] != "svc-a" || p.tags["type"] != "RPC" {
		t.Fatalf("tags = %+v, want service=svc-a type=RPC", p.tags)
	}
	if p.tags["name"] != "/backend.svc.Svc/Method" {
		t.Fatalf("name = %q, want FullMethod fallback", p.tags["name"])
	}
	if p.tags["status"] != SUCCESS {
		t.Fatalf("status = %q, want %q", p.tags["status"], SUCCESS)
	}
	if p.fields["grpc_code"] != "OK" {
		t.Fatalf("grpc_code = %v, want OK", p.fields["grpc_code"])
	}
}

func TestUnaryServerInterceptorError(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	iv := UnaryServerInterceptor(nil)
	_, err := iv(context.Background(), nil, unaryInfo(), func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("db timeout")
	})
	if err == nil || err.Error() != "db timeout" {
		t.Fatalf("err = %v, want db timeout passthrough", err)
	}

	p := drainPoint(t, ch)
	if p.tags["status"] != ERROR {
		t.Fatalf("status = %q, want %q", p.tags["status"], ERROR)
	}
	if p.tags["type"] != "RPC" {
		t.Fatalf("type = %q, want RPC", p.tags["type"])
	}
	if p.fields["error"] != "db timeout" {
		t.Fatalf("error = %v, want db timeout", p.fields["error"])
	}
	if p.fields["grpc_code"] != "Unknown" {
		t.Fatalf("grpc_code = %v, want Unknown for plain error", p.fields["grpc_code"])
	}
}

func TestUnaryServerInterceptorNameFn(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	t.Run("custom name wins", func(t *testing.T) {
		iv := UnaryServerInterceptor(func(*grpc.UnaryServerInfo) string { return "Svc/Method" })
		_, _ = iv(context.Background(), nil, unaryInfo(), func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		if p := drainPoint(t, ch); p.tags["name"] != "Svc/Method" {
			t.Fatalf("name = %q, want normalized name", p.tags["name"])
		}
	})

	t.Run("empty nameFn result falls back to FullMethod", func(t *testing.T) {
		iv := UnaryServerInterceptor(func(*grpc.UnaryServerInfo) string { return "" })
		_, _ = iv(context.Background(), nil, unaryInfo(), func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		if p := drainPoint(t, ch); p.tags["name"] != "/backend.svc.Svc/Method" {
			t.Fatalf("name = %q, want FullMethod fallback", p.tags["name"])
		}
	})
}

func TestUnaryServerInterceptorPanic(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	iv := UnaryServerInterceptor(nil)
	func() {
		defer func() {
			if rec := recover(); rec != "boom" {
				t.Fatalf("panic = %v, want boom re-panic", rec)
			}
		}()
		_, _ = iv(context.Background(), nil, unaryInfo(), func(ctx context.Context, req any) (any, error) {
			panic("boom")
		})
		t.Fatal("should have panicked")
	}()

	p := drainPoint(t, ch)
	if p.tags["status"] != ERROR {
		t.Fatalf("status = %q, want %q after panic", p.tags["status"], ERROR)
	}
}

func TestUnaryServerInterceptorDisabledPassthrough(t *testing.T) {
	resetCat(t) // 未 Init
	ch := capturePoints(t)

	iv := UnaryServerInterceptor(nil)
	resp, err := iv(context.Background(), "req", unaryInfo(), func(ctx context.Context, req any) (any, error) {
		if ctx.Value(cons.TraceIdKey) == "" {
			t.Error("trace id should still be injected when disabled")
		}
		return "resp", nil
	})
	if err != nil || resp != "resp" {
		t.Fatalf("resp = %v, err = %v, want passthrough", resp, err)
	}
	assertNoPoint(t, ch)

	// 失败路径同样放行
	_, err = iv(context.Background(), nil, unaryInfo(), func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("x")
	})
	if err == nil {
		t.Fatal("err should passthrough when disabled")
	}
	assertNoPoint(t, ch)
}
