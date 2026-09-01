package interceptors

import (
	"context"
	"log/slog"
	"time"

	"github.com/hookkster/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryLogging(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		l := log.With(slog.String("method", info.FullMethod))

		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		
		code := status.Code(err)

		attrs := []any{
			slog.Duration("duration", duration),
			slog.String("code", code.String()),
		}

		if err != nil {
			attrs = append(attrs, sl.Err(err))
		}

		switch {
		case code == codes.OK:
			l.Info("request handled", attrs...)
		case serverFault(code):
			l.Error("request failed", attrs...)
		default:
			l.Warn("request rejected", attrs...)
		}

		return resp, err
	}
}

func serverFault(code codes.Code) bool {
	switch code {
	case codes.Unknown, codes.Internal, codes.DeadlineExceeded, codes.Unavailable, codes.DataLoss:
		return true
	default:
		return false
	}
}
