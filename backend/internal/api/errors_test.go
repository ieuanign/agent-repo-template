package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/apperr"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

type failureBody struct {
	Meta struct {
		Path string `json:"path"`
	} `json:"meta"`
	Status  int             `json:"status"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Payload json.RawMessage `json:"payload"`
}

func TestErrorMapping(t *testing.T) {
	locked := apperr.New(apperr.Conflict, "PROJECTS_PROJECT_LOCKED", "The project is locked")
	const generic = "An internal error occurred."
	tests := []struct {
		name        string
		err         error
		path        string
		wantStatus  int
		wantCode    string
		wantMessage string
		notInBody   string
		wantLogged  bool
	}{
		{name: "not found", err: apperr.NotFound, wantStatus: 404, wantCode: "NOT_FOUND", wantMessage: apperr.NotFound.Message()},
		{name: "unauthorized", err: apperr.Unauthorized, wantStatus: 401, wantCode: "UNAUTHORIZED", wantMessage: apperr.Unauthorized.Message()},
		{name: "forbidden", err: apperr.Forbidden, wantStatus: 403, wantCode: "FORBIDDEN", wantMessage: apperr.Forbidden.Message()},
		{name: "validation", err: apperr.Validation, wantStatus: 400, wantCode: "VALIDATION", wantMessage: apperr.Validation.Message()},
		{name: "conflict", err: apperr.Conflict, wantStatus: 409, wantCode: "CONFLICT", wantMessage: apperr.Conflict.Message()},
		{name: "rate limited", err: apperr.RateLimited, wantStatus: 429, wantCode: "RATE_LIMITED", wantMessage: apperr.RateLimited.Message()},
		{name: "timeout", err: apperr.Timeout, wantStatus: 504, wantCode: "TIMEOUT", wantMessage: apperr.Timeout.Message()},
		{name: "internal", err: apperr.Internal, wantStatus: 500, wantCode: "INTERNAL", wantMessage: generic, wantLogged: true},
		{name: "refined", err: locked, wantStatus: 409, wantCode: "PROJECTS_PROJECT_LOCKED", wantMessage: "The project is locked"},
		{
			name: "wrapped refined", err: fmt.Errorf("handling: %w", fmt.Errorf("approving the estimate: %w", locked)),
			wantStatus: 409, wantCode: "PROJECTS_PROJECT_LOCKED", wantMessage: "The project is locked", notInBody: "approving",
		},
		{
			name: "plain error", err: errors.New("dial postgres://admin:hunter2@db"),
			wantStatus: 500, wantCode: "INTERNAL", wantMessage: generic, notInBody: "hunter2", wantLogged: true,
		},
		{name: "unknown kind", err: apperr.Kind{}, wantStatus: 500, wantCode: "INTERNAL", wantMessage: generic, wantLogged: true},
		{name: "unknown path", path: "/api/does-not-exist", wantStatus: 404, wantCode: "NOT_FOUND", wantMessage: apperr.NotFound.Message()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if path == "" {
				path = "/api/health"
			}
			core, logs := observer.New(zapcore.DebugLevel)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set(logger.RequestIDHeader, "req-42")
			rec := httptest.NewRecorder()
			NewRouter(stub{err: tt.err}, liveSpec(t), zap.New(core)).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q", ct)
			}
			var got failureBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if got.Status != tt.wantStatus || got.Code != tt.wantCode {
				t.Errorf("envelope status, code = %d, %q; want %d, %q", got.Status, got.Code, tt.wantStatus, tt.wantCode)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			if string(got.Payload) != "null" || got.Meta.Path != path {
				t.Errorf("payload = %s, meta.path = %q", got.Payload, got.Meta.Path)
			}
			if tt.notInBody != "" && strings.Contains(rec.Body.String(), tt.notInBody) {
				t.Errorf("body leaks %q: %s", tt.notInBody, rec.Body)
			}

			entries := logs.All()
			if !tt.wantLogged {
				if len(entries) != 0 {
					t.Fatalf("got %d log entries, want none", len(entries))
				}
				return
			}
			if len(entries) != 1 || entries[0].Level != zapcore.ErrorLevel {
				t.Fatalf("got %v, want one error-level entry", entries)
			}
			fields := entries[0].ContextMap()
			if fields["request_id"] != "req-42" {
				t.Errorf("logged request_id = %v", fields["request_id"])
			}
			if tt.notInBody != "" && !strings.Contains(fmt.Sprint(fields["error"]), tt.notInBody) {
				t.Errorf("logged error %v lacks %q", fields["error"], tt.notInBody)
			}
		})
	}
}

// errorCodeDrift compares the spec's ErrorCode enum with the codes of the kinds statuses maps.
func errorCodeDrift(t *testing.T, spec *openapi3.T) (missing, unknown []string) {
	t.Helper()
	ref, ok := spec.Components.Schemas["ErrorCode"]
	if !ok || ref.Value == nil {
		t.Fatal("spec has no ErrorCode schema")
	}
	listed := map[string]bool{}
	for _, v := range ref.Value.Enum {
		listed[fmt.Sprint(v)] = true
	}
	mapped := map[string]bool{}
	for kind := range statuses {
		mapped[kind.Code()] = true
		if !listed[kind.Code()] {
			missing = append(missing, kind.Code())
		}
	}
	for code := range listed {
		if !mapped[code] {
			unknown = append(unknown, code)
		}
	}
	slices.Sort(missing)
	slices.Sort(unknown)
	return missing, unknown
}

func TestErrorCodeEnumMatchesStatuses(t *testing.T) {
	tests := []struct {
		name        string
		edit        func(enum []any) []any
		wantMissing []string
		wantUnknown []string
	}{
		{name: "live spec", edit: func(e []any) []any { return e }},
		{
			name:        "code removed",
			edit:        func(e []any) []any { return slices.DeleteFunc(e, func(v any) bool { return v == "TIMEOUT" }) },
			wantMissing: []string{"TIMEOUT"},
		},
		{
			name:        "unknown code",
			edit:        func(e []any) []any { return append(e, "TEAPOT") },
			wantUnknown: []string{"TEAPOT"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := liveSpec(t)
			if ref, ok := spec.Components.Schemas["ErrorCode"]; ok && ref.Value != nil {
				ref.Value.Enum = tt.edit(ref.Value.Enum)
			}
			missing, unknown := errorCodeDrift(t, spec)
			if !slices.Equal(missing, tt.wantMissing) {
				t.Errorf("kinds missing from ErrorCode = %v, want %v", missing, tt.wantMissing)
			}
			if !slices.Equal(unknown, tt.wantUnknown) {
				t.Errorf("ErrorCode values with no kind = %v, want %v", unknown, tt.wantUnknown)
			}
		})
	}
}
