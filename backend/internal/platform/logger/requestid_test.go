package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
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
			r := gin.New()
			r.ContextWithFallback = true
			r.Use(RequestID(zap.New(core)))
			var ctxID string
			var ctxOK bool
			// Handlers receive the *gin.Context as their context.Context.
			r.GET("/x", func(c *gin.Context) {
				FromContext(c).Info("handled")
				ctxID, ctxOK = RequestIDFromContext(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			if tt.incoming != "" {
				req.Header.Set(RequestIDHeader, tt.incoming)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			got := rec.Header().Get(RequestIDHeader)
			if got == "" || (tt.incoming != "" && got != tt.incoming) {
				t.Fatalf("response %s = %q, incoming %q", RequestIDHeader, got, tt.incoming)
			}
			entries := logs.FilterMessage("handled").All()
			if len(entries) != 1 {
				t.Fatalf("got %d log entries, want 1", len(entries))
			}
			if id := entries[0].ContextMap()["request_id"]; id != got {
				t.Fatalf("logged request_id = %v, want %q", id, got)
			}
			if !ctxOK || ctxID != got {
				t.Fatalf("RequestIDFromContext = %q, %v; want %q, true", ctxID, ctxOK, got)
			}
		})
	}
}

func TestFromContextWithoutLogger(t *testing.T) {
	FromContext(t.Context()).Info("must not panic")
}

func TestRequestIDFromContextWithoutID(t *testing.T) {
	if id, ok := RequestIDFromContext(t.Context()); ok || id != "" {
		t.Fatalf("RequestIDFromContext = %q, %v; want \"\", false", id, ok)
	}
}
