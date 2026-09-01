package interceptors

import (
	"context"
	"log/slog"
	"time"

	"github.com/hookkster/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const TraceIDKey = "x-trace-id"

func UnaryLogging(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		traceID := TraceIDFromContext(ctx)
		if traceID == "" {
			traceID = sl.NewTraceID()
		}

		ctx = sl.WithTraceID(ctx, traceID)

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
			l.InfoContext(ctx, "request handled", attrs...)
		case serverFault(code):
			l.ErrorContext(ctx, "request failed", attrs...)
		default:
			l.WarnContext(ctx, "request rejected", attrs...)
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

func TraceIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(TraceIDKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
