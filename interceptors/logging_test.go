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

func TestUnaryLoggingCallsHandler(t *testing.T) {
	called := false
	handler := func(_ context.Context, _ any) (any, error) {
		called = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	interceptor := interceptors.UnaryLogging(log)

	resp, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}

	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

func TestUnaryLoggingWritesRecord(t *testing.T) {
	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	interceptor := interceptors.UnaryLogging(log)

	_, _ = interceptor(context.Background(), nil, info, handler)

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &record); err != nil {
		t.Fatalf("invalid json: %v (%q)", err, buf.String())
	}

	if record["method"] != "/auth.v1.AuthService/SendOTP" {
		t.Errorf("method = %v, want /auth.v1.AuthService/SendOTP", record["method"])
	}

	d, ok := record["duration"].(float64)
	if !ok || d <= 0 {
		t.Errorf("duration = %v, want positive number", record["duration"])
	}

	if record["code"] != "OK" {
		t.Errorf("code = %v, want OK", record["code"])
	}
}

func TestUnaryLoggingRecordsErrorCode(t *testing.T) {
	wantErr := status.Error(codes.NotFound, "not found")

	handler := func(_ context.Context, _ any) (any, error) {
		return nil, wantErr
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/SendOTP"}

	buf := &bytes.Buffer{}
	log := sl.NewLogger(sl.Config{Env: sl.EnvProd, Output: buf})

	interceptor := interceptors.UnaryLogging(log)

	_, err := interceptor(context.Background(), nil, info, handler)
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &record); err != nil {
		t.Fatalf("invalid json: %v (%q)", err, buf.String())
	}

	if record["code"] != "NotFound" {
		t.Errorf("code = %v, want NotFound", record["code"])
	}

	if record["level"] != "WARN" {
		t.Errorf("level = %v, want WARN", record["level"])
	}
}
