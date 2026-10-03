package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

func docsRouter(t *testing.T, env config.AppEnv) *gin.Engine {
	t.Helper()
	specJSON, err := openapi.GetSpecJSON()
	if err != nil {
		t.Fatal(err)
	}
	return NewRouter(NewServer("1.0.0", env, fakePinger{}), liveSpec(t), zap.NewNop(), Docs(env, specJSON))
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

var assetRef = regexp.MustCompile(`(?:src|href)="([^"]*)"`)

func TestDocsPage(t *testing.T) {
	for _, env := range []config.AppEnv{config.Dev, config.Staging} {
		for _, path := range []string{"/api/docs", "/api/docs/"} {
			t.Run(string(env)+" "+path, func(t *testing.T) {
				r := docsRouter(t, env)
				rec := get(r, path)
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", rec.Code)
				}
				if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
					t.Fatalf("Content-Type = %q", ct)
				}
				page := rec.Body.String()
				for _, want := range []string{"/api/openapi.json", "validatorUrl: null"} {
					if !strings.Contains(page, want) {
						t.Errorf("page lacks %q", want)
					}
				}
				refs := assetRef.FindAllStringSubmatch(page, -1)
				if len(refs) == 0 {
					t.Fatal("page references no assets")
				}
				for _, m := range refs {
					ref := m[1]
					if !strings.HasPrefix(ref, "/api/docs/") {
						t.Errorf("asset %q is not a same-origin path under /api/docs/", ref)
						continue
					}
					if a := get(r, ref); a.Code != http.StatusOK || a.Body.Len() == 0 {
						t.Errorf("asset %q: status %d, %d bytes", ref, a.Code, a.Body.Len())
					}
				}
			})
		}
	}
}

func TestDocsSpec(t *testing.T) {
	want, err := openapi3.NewLoader().LoadFromFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, env := range []config.AppEnv{config.Dev, config.Staging} {
		t.Run(string(env), func(t *testing.T) {
			rec := get(docsRouter(t, env), "/api/openapi.json")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Fatalf("Content-Type = %q", ct)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
				t.Fatal(err)
			}
			if _, ok := raw["payload"]; ok {
				t.Fatal("spec is wrapped in the envelope")
			}
			got, err := openapi3.NewLoader().LoadFromData(rec.Body.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Servers) == 0 || got.Servers[0].URL != "/api" {
				t.Errorf("servers = %v, want [/api]", got.Servers)
			}
			if g, w := operations(got), operations(want); !slices.Equal(g, w) {
				t.Errorf("operations = %v, want %v", g, w)
			}
		})
	}
}

// operations lists "METHOD path operationId", sorted.
func operations(spec *openapi3.T) []string {
	var out []string
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			out = append(out, method+" "+path+" "+op.OperationID)
		}
	}
	slices.Sort(out)
	return out
}

func TestDocsHiddenInProduction(t *testing.T) {
	r := docsRouter(t, config.Production)
	for _, path := range []string{"/api/docs", "/api/docs/swagger-ui-bundle.js", "/api/openapi.json", "/api/does-not-exist"} {
		t.Run(path, func(t *testing.T) {
			rec := get(r, path)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			var got failureBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if got.Code != "NOT_FOUND" {
				t.Errorf("code = %q, want NOT_FOUND", got.Code)
			}
		})
	}
}
