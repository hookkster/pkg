package interceptors

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryRecovery(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(
					ctx,
					"panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)

				err = status.Error(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}
