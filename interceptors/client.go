package interceptors

import (
	"context"

	"github.com/hookkster/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryClientTraceID() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		traceID := sl.TraceID(ctx)

		if traceID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, TraceIDKey, traceID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
