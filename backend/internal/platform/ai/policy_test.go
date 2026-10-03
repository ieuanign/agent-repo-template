package ai

import (
	"context"
	"slices"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aiv1 "github.com/ieuanign/agent-repo-template/backend/gen/ai/v1"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

func TestDefaultDeadline(t *testing.T) {
	tests := []struct {
		name string
		own  time.Duration // 0: the caller sets no deadline
		want time.Duration
	}{
		{"none set gets 10s", 0, 10 * time.Second},
		{"shorter own deadline kept", 2 * time.Second, 2 * time.Second},
		{"longer own deadline kept", 30 * time.Second, 30 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeInfo{}
			conn := dialFake(t, fake)
			ctx := t.Context()
			if tt.own > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.own)
				defer cancel()
			}

			start := time.Now()
			if _, err := aiv1.NewInfoServiceClient(conn).GetInfo(ctx, &aiv1.GetInfoRequest{}); err != nil {
				t.Fatalf("GetInfo: %v", err)
			}

			calls := fake.seen()
			if len(calls) != 1 {
				t.Fatalf("got %d calls; want 1", len(calls))
			}
			// The server rebuilds the deadline from grpc-timeout on arrival, so it lands just after start+want.
			want := start.Add(tt.want)
			if d := calls[0].deadline; d.Before(want.Add(-time.Second)) || d.After(want.Add(time.Second)) {
				t.Fatalf("deadline %s after start; want about %s", d.Sub(start), tt.want)
			}
		})
	}
}

func TestRequestIDMetadata(t *testing.T) {
	tests := []struct {
		name string
		ctx  func(context.Context) context.Context
		want []string // nil: the key is absent
	}{
		{"ID in context is sent", func(ctx context.Context) context.Context { return logger.WithRequestID(ctx, "req-123") }, []string{"req-123"}},
		{"no ID sends no key", func(ctx context.Context) context.Context { return ctx }, nil},
		{"non-printable ID is not sent", func(ctx context.Context) context.Context { return logger.WithRequestID(ctx, "café") }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeInfo{}
			conn := dialFake(t, fake)

			if _, err := aiv1.NewInfoServiceClient(conn).GetInfo(tt.ctx(t.Context()), &aiv1.GetInfoRequest{}); err != nil {
				t.Fatalf("GetInfo: %v", err)
			}

			calls := fake.seen()
			if len(calls) != 1 {
				t.Fatalf("got %d calls; want 1", len(calls))
			}
			if got := calls[0].requestID; !slices.Equal(got, tt.want) {
				t.Fatalf("x-request-id %q; want %q", got, tt.want)
			}
		})
	}
}

func TestRetryOnlyUnavailable(t *testing.T) {
	tests := []struct {
		name      string
		script    []codes.Code
		wantCode  codes.Code
		wantCalls int
	}{
		{"UNAVAILABLE twice then success", []codes.Code{codes.Unavailable, codes.Unavailable}, codes.OK, 3},
		{"DEADLINE_EXCEEDED is not retried", []codes.Code{codes.DeadlineExceeded}, codes.DeadlineExceeded, 1},
		{"RESOURCE_EXHAUSTED is not retried", []codes.Code{codes.ResourceExhausted}, codes.ResourceExhausted, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeInfo{script: tt.script}
			conn := dialFake(t, fake)

			_, err := aiv1.NewInfoServiceClient(conn).GetInfo(t.Context(), &aiv1.GetInfoRequest{})

			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("code %s; want %s (err %v)", got, tt.wantCode, err)
			}
			if got := len(fake.seen()); got != tt.wantCalls {
				t.Fatalf("got %d calls; want %d", got, tt.wantCalls)
			}
		})
	}
}
