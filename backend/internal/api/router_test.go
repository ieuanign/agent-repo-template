package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
	}{
		{"echoes incoming", "abc-123"},
		{"assigns when absent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			impl := stub{resp: statusOnly(200), call: func(ctx context.Context) { logger.FromContext(ctx).Info("handled") }}
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			if tt.incoming != "" {
				req.Header.Set(logger.RequestIDHeader, tt.incoming)
			}
			rec := httptest.NewRecorder()
			NewRouter(impl, liveSpec(t), zap.New(core)).ServeHTTP(rec, req)

			got := rec.Header().Get(logger.RequestIDHeader)
			if got == "" || (tt.incoming != "" && got != tt.incoming) {
				t.Fatalf("response %s = %q, incoming %q", logger.RequestIDHeader, got, tt.incoming)
			}
			entries := logs.FilterMessage("handled").All()
			if len(entries) != 1 {
				t.Fatalf("got %d log entries, want 1", len(entries))
			}
			if id := entries[0].ContextMap()["request_id"]; id != got {
				t.Fatalf("logged request_id = %v, want %q", id, got)
			}
		})
	}
}
