package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunHealthcheck(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		closed  bool
		wantErr bool
	}{
		{"ready", http.StatusOK, false, false},
		{"not ready", http.StatusServiceUnavailable, false, true},
		{"closed port", 0, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.Path
				w.WriteHeader(tt.status)
			}))
			_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
			if tt.closed {
				srv.Close()
			} else {
				defer srv.Close()
			}
			environ := []string{"APP_ENV=dev", "PORT=" + port, "POSTGRES_HOST=postgres",
				"POSTGRES_USER=app", "POSTGRES_PASSWORD=pw", "POSTGRES_DB=template", "AI_HOST=10.255.255.1"}
			var out bytes.Buffer
			err := run(context.Background(), []string{"server", "healthcheck"}, environ, &out, none, mustNotOpen(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("run healthcheck: err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.closed && path != "/api/health/ready" {
				t.Errorf("requested %q, want /api/health/ready", path)
			}
			if out.Len() != 0 {
				t.Errorf("healthcheck wrote startup output:\n%s", out.String())
			}
		})
	}
}

func TestRunHealthcheckConfigError(t *testing.T) {
	if err := run(context.Background(), []string{"server", "healthcheck"}, []string{"PORT=x"}, &bytes.Buffer{}, none, mustNotOpen(t)); err == nil {
		t.Fatal("run healthcheck: want error, got nil")
	}
}
