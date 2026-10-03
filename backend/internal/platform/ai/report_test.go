package ai

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	aiv1 "github.com/ieuanign/agent-repo-template/backend/gen/ai/v1"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

// fakeInfo answers with script's codes in turn, then with getInfo, or success when that is nil.
type fakeInfo struct {
	aiv1.UnimplementedInfoServiceServer
	getInfo func(context.Context) (*aiv1.GetInfoResponse, error)
	script  []codes.Code

	mu    sync.Mutex
	calls []received
}

type received struct {
	deadline  time.Time
	requestID []string
}

func (f *fakeInfo) GetInfo(ctx context.Context, _ *aiv1.GetInfoRequest) (*aiv1.GetInfoResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	deadline, _ := ctx.Deadline()
	f.mu.Lock()
	n := len(f.calls)
	f.calls = append(f.calls, received{deadline: deadline, requestID: md.Get("x-request-id")})
	f.mu.Unlock()
	if n < len(f.script) {
		return nil, status.Error(f.script[n], "scripted")
	}
	if f.getInfo != nil {
		return f.getInfo(ctx)
	}
	return &aiv1.GetInfoResponse{}, nil
}

func (f *fakeInfo) seen() []received {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]received(nil), f.calls...)
}

// dialFake serves srv on bufconn and connects through New; the IP-literal host skips DNS.
func dialFake(t *testing.T, srv aiv1.InfoServiceServer) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	aiv1.RegisterInfoServiceServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)
	conn, err := New(config.AI{Host: "127.0.0.1", Port: 50051},
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func observed() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return zap.New(core), logs
}

func TestReportLogsServiceAndVersion(t *testing.T) {
	conn := dialFake(t, &fakeInfo{getInfo: func(context.Context) (*aiv1.GetInfoResponse, error) {
		return &aiv1.GetInfoResponse{Service: "ai", Version: "1.2.3"}, nil
	}})
	log, logs := observed()

	Report(context.Background(), conn, log)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("got %d log entries; want 1", len(entries))
	}
	e := entries[0]
	if e.Level != zapcore.InfoLevel {
		t.Fatalf("level %s; want info", e.Level)
	}
	fields := e.ContextMap()
	if fields["service"] != "ai" || fields["version"] != "1.2.3" {
		t.Fatalf("fields %v; want service=ai version=1.2.3", fields)
	}
}

// onlyWarn fails unless logs hold exactly one Warn entry whose error contains want.
func onlyWarn(t *testing.T, logs *observer.ObservedLogs, want string) {
	t.Helper()
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("got %d log entries; want 1", len(entries))
	}
	e := entries[0]
	if e.Level != zapcore.WarnLevel {
		t.Fatalf("level %s; want warn", e.Level)
	}
	if got, _ := e.ContextMap()["error"].(string); !strings.Contains(got, want) {
		t.Fatalf("error field %q; want it to contain %q", got, want)
	}
}

func TestReportWarnsOnStatusError(t *testing.T) {
	conn := dialFake(t, &fakeInfo{getInfo: func(context.Context) (*aiv1.GetInfoResponse, error) {
		return nil, status.Error(codes.Unavailable, "ai is warming up")
	}})
	log, logs := observed()

	Report(context.Background(), conn, log)

	onlyWarn(t, logs, "ai is warming up")
}

func TestReportGivesUpAtTheParentDeadline(t *testing.T) {
	conn := dialFake(t, &fakeInfo{getInfo: func(ctx context.Context) (*aiv1.GetInfoResponse, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}})
	log, logs := observed()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	Report(ctx, conn, log)

	if d := time.Since(start); d > time.Second {
		t.Fatalf("Report took %s; want about the 100ms parent deadline", d)
	}
	onlyWarn(t, logs, "DeadlineExceeded")
}
