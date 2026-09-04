package interceptors_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hookkster/pkg/interceptors"
	"github.com/hookkster/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryRecoveryConvertsPanic(t *testing.T) {
	handler := func(_ context.Context, _ any) (any, error) {
		panic("test panic")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	interceptor := interceptors.UnaryRecovery(log)

	resp, err := interceptor(context.Background(), nil, info, handler)

	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}

	if status.Code(err) != codes.Internal {
		t.Errorf("code = %v, want Internal", status.Code(err))
	}

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &record); err != nil {
		t.Fatalf("invalid json: %v (%q)", err, buf.String())
	}

	if record["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", record["level"])
	}
}

func TestUnaryRecoveryPassesErrorsThrough(t *testing.T) {
	wantErr := status.Error(codes.NotFound, "not found")

	handler := func(_ context.Context, _ any) (any, error) {
		return nil, wantErr
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	interceptor := interceptors.UnaryRecovery(log)

	_, err := interceptor(context.Background(), nil, info, handler)

	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}

	if buf.Len() != 0 {
		t.Errorf("recovery logged on a normal path: %q", buf.String())
	}
}

func TestInterceptorChainSharesTraceID(t *testing.T) {
	panicHandler := func(_ context.Context, _ any) (any, error) {
		panic("test panic")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	logging := interceptors.UnaryLogging(log)
	recovery := interceptors.UnaryRecovery(log)

	wrapped := func(ctx context.Context, req any) (any, error) {
		return recovery(ctx, req, info, panicHandler)
	}

	_, _ = logging(context.Background(), nil, info, wrapped)

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))

	if len(lines) != 2 {
		t.Fatalf("logs = %v, want 2", len(lines))
	}

	var recordRecovery map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(lines[0]), &recordRecovery); err != nil {
		t.Fatalf("invalid json: %v (%q)", err, buf.String())
	}

	var recordLogging map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(lines[1]), &recordLogging); err != nil {
		t.Fatalf("invalid json: %v (%q)", err, buf.String())
	}

	id, ok := recordRecovery["trace_id"].(string)
	if !ok || id == "" {
		t.Errorf("recovery trace_id = %v, want non-empty string", recordRecovery["trace_id"])
	}

	if recordLogging["trace_id"] != id {
		t.Errorf("logging trace_id = %v, want %v", recordLogging["trace_id"], id)
	}

	if recordLogging["trace_id"] != recordRecovery["trace_id"] {
		t.Errorf("trace_id logging = %v, recovery = %v", recordLogging["trace_id"], recordRecovery["trace_id"])
	}

	if recordLogging["code"] != "Internal" {
		t.Errorf("code = %v want Internal", recordLogging["code"])
	}
}
