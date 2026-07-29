// [[ when (modeIs "http") ]]
package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"[[.ModulePath]]/api"
	httptransport "[[.ModulePath]]/internal/transport/http"
)

// specDoc is the slice of the OpenAPI document these tests reason about.
type specDoc struct {
	OpenAPI string                       `yaml:"openapi"`
	Info    struct{ Title, Version string } `yaml:"info"`
	Paths   map[string]map[string]any    `yaml:"paths"`
}

func loadSpec(t *testing.T) specDoc {
	t.Helper()
	var doc specDoc
	if err := yaml.Unmarshal(api.SpecYAML, &doc); err != nil {
		t.Fatalf("parse embedded spec: %v", err)
	}
	return doc
}

func TestSpecIsWellFormed(t *testing.T) {
	doc := loadSpec(t)
	if !strings.HasPrefix(doc.OpenAPI, "3.") {
		t.Fatalf("openapi = %q, want a 3.x document", doc.OpenAPI)
	}
	if doc.Info.Title == "" || doc.Info.Version == "" {
		t.Fatalf("info.title/version must be set, got %+v", doc.Info)
	}
	if len(doc.Paths) == 0 {
		t.Fatal("spec documents no paths")
	}
}

func TestSpecConvertsToJSON(t *testing.T) {
	raw, err := api.SpecJSON()
	if err != nil {
		t.Fatalf("SpecJSON: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("SpecJSON output is not valid JSON: %v", err)
	}
	if _, ok := doc["paths"]; !ok {
		t.Fatal("JSON spec has no paths key")
	}
}

// TestSpecDocumentsOnlyRealRoutes catches a spec that has drifted ahead of the
// code: every documented operation must actually be routed.
func TestSpecDocumentsOnlyRealRoutes(t *testing.T) {
	router := newRouter(t)

	for path, ops := range loadSpec(t).Paths {
		for method := range ops {
			m := strings.ToUpper(method)
			t.Run(m+" "+path, func(t *testing.T) {
				// Send a minimal valid body so POSTs reach the handler rather
				// than failing to decode; only routing is under test here.
				var body *strings.Reader
				req := httptest.NewRequest(m, path, nil)
				if m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch {
					body = strings.NewReader(`{"message":"routing probe"}`)
					req = httptest.NewRequest(m, path, body)
					req.Header.Set("Content-Type", "application/json")
				}

				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				switch rec.Code {
				case http.StatusNotFound:
					t.Fatalf("spec documents %s %s but nothing is routed there", m, path)
				case http.StatusMethodNotAllowed:
					t.Fatalf("spec documents %s %s but that method is not allowed", m, path)
				}
			})
		}
	}
}

// TestSpecDocumentsEveryPublicRoute catches code that has drifted ahead of the
// spec. net/http's ServeMux cannot enumerate its patterns, so the service's
// public surface is listed here on purpose: adding a route means adding it to
// this list, which fails until the spec documents it too.
func TestSpecDocumentsEveryPublicRoute(t *testing.T) {
	publicRoutes := []struct{ method, path string }{
		{http.MethodGet, "/healthz"},
		{http.MethodPost, "/echo"},
	}

	paths := loadSpec(t).Paths
	for _, want := range publicRoutes {
		ops, ok := paths[want.path]
		if !ok {
			t.Errorf("route %s %s is served but the spec has no %q path", want.method, want.path, want.path)
			continue
		}
		if _, ok := ops[strings.ToLower(want.method)]; !ok {
			t.Errorf("route %s %s is served but the spec documents no %s operation for it",
				want.method, want.path, strings.ToLower(want.method))
		}
	}
}

func TestDocsEndpointsServeWhenEnabled(t *testing.T) {
	router := newRouterWith(t, httptransport.Options{MaxBodyBytes: 1 << 20, DocsEnabled: true})

	tests := map[string]struct {
		path        string
		contentType string
		contains    string
	}{
		"yaml spec": {httptransport.SpecPathYAML, "application/yaml", "openapi:"},
		"json spec": {httptransport.SpecPathJSON, "application/json", `"openapi"`},
		"ui page":   {httptransport.DocsPath, "text/html", "SwaggerUIBundle"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s = %d, want 200", tc.path, rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, tc.contentType) {
				t.Fatalf("Content-Type = %q, want %q", ct, tc.contentType)
			}
			if !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("body does not contain %q", tc.contains)
			}
		})
	}
}

// The UI must pin its assets with integrity hashes; an unpinned CDN script would
// let a compromised origin run arbitrary JS on the docs page.
func TestDocsPagePinsAssetsWithIntegrity(t *testing.T) {
	router := newRouterWith(t, httptransport.Options{MaxBodyBytes: 1 << 20, DocsEnabled: true})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, httptransport.DocsPath, nil))

	page := rec.Body.String()
	for _, want := range []string{`integrity="sha384-`, `crossorigin="anonymous"`} {
		if strings.Count(page, want) < 2 {
			t.Errorf("docs page should apply %s to both the CSS and JS tags", want)
		}
	}
	if strings.Contains(page, "petstore.swagger.io") {
		t.Error("docs page still points at the Swagger UI demo spec")
	}
}

func TestDocsEndpointsAbsentWhenDisabled(t *testing.T) {
	router := newRouterWith(t, httptransport.Options{MaxBodyBytes: 1 << 20, DocsEnabled: false})

	for _, path := range []string{
		httptransport.SpecPathYAML, httptransport.SpecPathJSON, httptransport.DocsPath,
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d with docs disabled, want 404", path, rec.Code)
		}
	}
}

// Disabling docs must not disable the API itself.
func TestAPIStillServesWhenDocsDisabled(t *testing.T) {
	router := newRouterWith(t, httptransport.Options{MaxBodyBytes: 1 << 20, DocsEnabled: false})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("GET /healthz = %d, want 204", rec.Code)
	}
}
