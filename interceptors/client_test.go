package interceptors_test

import (
	"context"
	"testing"

	"github.com/hookkster/pkg/interceptors"
	"github.com/hookkster/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientTraceIDPropagates(t *testing.T) {
	var got []string

	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		got = md.Get(interceptors.TraceIDKey)

		return nil
	}

	ctx := sl.WithTraceID(context.Background(), "abc123")

	err := interceptors.UnaryClientTraceID()(ctx, "/test/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) == 0 {
		t.Fatalf("got %v values, want 1", len(got))
	}

	if got[0] != "abc123" {
		t.Errorf("trace_id = %v want abc123", got[0])
	}
}

func TestUnaryClientTraceIDSkipsWhenAbsent(t *testing.T) {
	var got []string

	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		got = md.Get(interceptors.TraceIDKey)

		return nil
	}

	err := interceptors.UnaryClientTraceID()(context.Background(), "/test/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %v values, want 0", len(got))
	}
}
