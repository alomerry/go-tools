package cat

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor 返回 gRPC Unary Server 拦截器：为每个 RPC 创建事务、
// 注入 trace id、按 handler 返回的 err 置状态并产出 transaction 点位。
// nameFn 返回事务名（建议传归一化函数，避免原始 FullMethod 参数化不足时的高基数）；
// 为 nil 或返回空时回退 info.FullMethod。与 HTTPMiddleware 的 nameFn 语义一致。
// 未初始化时静默放行（事务为空实现），但 trace id 仍注入 ctx，保证链路不断链。
func UnaryServerInterceptor(nameFn func(*grpc.UnaryServerInfo) string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tx, nctx := NewTransactionWithCtx(ctx, transactionTypeRPC, serverName(info, nameFn))

		defer func() {
			if rec := recover(); rec != nil {
				tx.SetStatus(ERROR)
				tx.Complete()
				panic(rec)
			}
		}()

		resp, err := handler(nctx, req)
		if err != nil {
			tx.SetStatus(ERROR)
			tx.AddData("error", err.Error())
		}
		tx.AddData("grpc_code", status.Code(err).String())
		tx.Complete()
		return resp, err
	}
}

func serverName(info *grpc.UnaryServerInfo, nameFn func(*grpc.UnaryServerInfo) string) string {
	var got string
	if nameFn != nil {
		got = nameFn(info)
	}
	fallback := ""
	if info != nil {
		fallback = info.FullMethod
	}
	return resolveName(got, fallback)
}
