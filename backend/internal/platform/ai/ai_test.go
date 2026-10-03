package ai

import (
	"testing"
	"time"

	"google.golang.org/grpc/connectivity"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

func TestNewDoesNotDial(t *testing.T) {
	// 10.255.255.1 is unroutable, so a dial would hang rather than fail fast.
	start := time.Now()
	conn, err := New(config.AI{Host: "10.255.255.1", Port: 50051})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer conn.Close()
	if d := time.Since(start); d > time.Second {
		t.Fatalf("New took %s; it must not dial", d)
	}
	if s := conn.GetState(); s != connectivity.Idle {
		t.Fatalf("state %s; want Idle", s)
	}
}
