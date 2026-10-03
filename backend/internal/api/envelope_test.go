package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
)

// stub answers GetHealth with whatever the test sets.
type stub struct {
	resp  openapi.GetHealthResponseObject
	err   error
	panic bool
	call  func(ctx context.Context)
}

func (s stub) GetHealth(ctx context.Context, _ openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	if s.panic {
		panic("secret panic detail")
	}
	if s.call != nil {
		s.call(ctx)
	}
	return s.resp, s.err
}

func (stub) GetHealthReady(context.Context, openapi.GetHealthReadyRequestObject) (openapi.GetHealthReadyResponseObject, error) {
	return openapi.GetHealthReady200Response{}, nil
}

// statusOnly writes a status and no body.
type statusOnly int

func (s statusOnly) VisitGetHealthResponse(w http.ResponseWriter) error {
	w.WriteHeader(int(s))
	return nil
}

// liveSpec is the embedded spec cmd/server validates against.
func liveSpec(t *testing.T) *openapi3.T {
	t.Helper()
	spec, err := openapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func serve(t *testing.T, impl openapi.StrictServerInterface, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	NewRouter(impl, liveSpec(t), zap.NewNop()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		impl       stub
		path       string
		wantStatus int
		wantBody   string // exact when bodyExact, else must be contained
		bodyExact  bool
		wantJSON   bool
		notInBody  string
	}{
		{name: "204 has no body", impl: stub{resp: statusOnly(204)}, path: "/api/health", wantStatus: 204, bodyExact: true},
		{name: "304 has no body", impl: stub{resp: statusOnly(304)}, path: "/api/health", wantStatus: 304, bodyExact: true},
		{name: "200 without body wraps null", impl: stub{resp: statusOnly(200)}, path: "/api/health", wantStatus: 200, wantBody: `"payload":null`, wantJSON: true},
		{name: "handler error hides its text", impl: stub{err: errors.New("secret db failure")}, path: "/api/health", wantStatus: 500, wantJSON: true, notInBody: "secret"},
		{name: "panic answers INTERNAL, hiding its detail", impl: stub{panic: true}, path: "/api/health", wantStatus: 500, wantBody: `"status":500,"code":"INTERNAL"`, wantJSON: true, notInBody: "secret"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, tt.impl, tt.path)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			body := rec.Body.String()
			if tt.bodyExact && body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if !tt.bodyExact && !strings.Contains(body, tt.wantBody) {
				t.Errorf("body %q lacks %q", body, tt.wantBody)
			}
			if tt.notInBody != "" && strings.Contains(body, tt.notInBody) {
				t.Errorf("body %q contains %q", body, tt.notInBody)
			}
			isJSON := strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json")
			if isJSON != tt.wantJSON {
				t.Errorf("Content-Type = %q, want JSON %v", rec.Header().Get("Content-Type"), tt.wantJSON)
			}
		})
	}
}
