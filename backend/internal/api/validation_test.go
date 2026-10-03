package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

func TestRequestValidation(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromFile("testdata/validation-spec.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
		wantCalled bool
	}{
		{name: "missing required query parameter", path: "/api/health", wantStatus: 400, wantCode: "VALIDATION"},
		{name: "required query parameter present", path: "/api/health?probe=x", wantStatus: 200, wantCalled: true},
		{name: "route missing from the spec", path: "/api/health/ready", wantStatus: 404, wantCode: "NOT_FOUND"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			impl := stub{resp: statusOnly(200), call: func(context.Context) { called = true }}
			rec := httptest.NewRecorder()
			NewRouter(impl, spec, zap.NewNop()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if called != tt.wantCalled {
				t.Errorf("stub called = %v, want %v", called, tt.wantCalled)
			}
			if tt.wantCode == "" {
				return
			}
			var got failureBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if got.Status != tt.wantStatus || got.Code != tt.wantCode || got.Message == "" || string(got.Payload) != "null" {
				t.Errorf("envelope = %+v; want status %d, code %q, a message, payload null", got, tt.wantStatus, tt.wantCode)
			}
		})
	}
}
