package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

func TestGetHealth(t *testing.T) {
	const version, commit = "9.9.9", "abc1234"
	for _, env := range []config.AppEnv{config.Dev, config.Staging, config.Production} {
		t.Run(string(env), func(t *testing.T) {
			// A failing pinger: liveness must not depend on it.
			r := NewRouter(NewServer(version, env, fakePinger{errors.New("down")}), liveSpec(t), zap.NewNop())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Fatalf("Content-Type = %q", ct)
			}
			if strings.Contains(rec.Body.String(), commit) {
				t.Fatalf("body leaks the commit: %s", rec.Body)
			}
			var top map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &top); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if len(top) != 5 {
				t.Fatalf("got %d top-level keys, want 5: %s", len(top), rec.Body)
			}
			var got struct {
				Meta struct {
					Path      string `json:"path"`
					Timestamp string `json:"timestamp"`
				} `json:"meta"`
				Status  int     `json:"status"`
				Code    *string `json:"code"`
				Message string  `json:"message"`
				Payload struct {
					Version     string `json:"version"`
					Environment string `json:"environment"`
				} `json:"payload"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Status != 200 || got.Code != nil || string(top["code"]) != "null" {
				t.Errorf("status = %d, code = %s; want 200, null", got.Status, top["code"])
			}
			if got.Message != "Backend is up." {
				t.Errorf("message = %q", got.Message)
			}
			if got.Payload.Version != version || got.Payload.Environment != string(env) {
				t.Errorf("payload = %+v", got.Payload)
			}
			if got.Meta.Path != "/api/health" {
				t.Errorf("meta.path = %q", got.Meta.Path)
			}
			if _, err := time.Parse(time.RFC3339, got.Meta.Timestamp); err != nil {
				t.Errorf("meta.timestamp %q: %v", got.Meta.Timestamp, err)
			}
		})
	}
}

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func TestGetHealthReady(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
	}{
		{"pinger succeeds", nil, http.StatusOK},
		{"pinger fails", errors.New("dial tcp postgres:5432: connection refused"), http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(NewServer("1.0.0", config.Dev, fakePinger{tt.pingErr}), liveSpec(t), zap.NewNop())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health/ready", nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			var top map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &top); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if string(top["status"]) != strconv.Itoa(tt.wantStatus) {
				t.Errorf("envelope status = %s", top["status"])
			}
			if string(top["code"]) != "null" || string(top["payload"]) != "null" {
				t.Errorf("code = %s, payload = %s; want null, null", top["code"], top["payload"])
			}
			if string(top["message"]) == `""` || top["message"] == nil {
				t.Errorf("message is empty: %s", rec.Body)
			}
			body := strings.ToLower(rec.Body.String())
			for _, leak := range []string{"connection refused", "postgres", "database"} {
				if strings.Contains(body, leak) {
					t.Errorf("body leaks %q: %s", leak, rec.Body)
				}
			}
		})
	}
}
