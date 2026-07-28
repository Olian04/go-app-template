// [[ when (modeIs "http") ]]
package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"[[.ModulePath]]/internal/config"
	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/logging"
	"[[.ModulePath]]/internal/observability/metrics"
	httptransport "[[.ModulePath]]/internal/transport/http"
)

func newRouter(t *testing.T) http.Handler {
	t.Helper()
	registry, err := metrics.NewRegistry(metrics.WithNamespace("test"))
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	cfg := config.HTTPSection{}.WithDefaults()
	return httptransport.Router(echo.NewService(), registry, cfg.MaxBodyBytes)
}

func TestRouterGeneratesRequestID(t *testing.T) {
	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if got := rec.Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}
	id := rec.Header().Get(logging.RequestIDHeader)
	if len(id) != 32 {
		t.Fatalf("%s = %q, want a generated 32-char hex id", logging.RequestIDHeader, id)
	}
}

func TestRouterPropagatesInboundRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(logging.RequestIDHeader, "trace-42")

	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, req)

	if got := rec.Header().Get(logging.RequestIDHeader); got != "trace-42" {
		t.Fatalf("%s = %q, want inbound id echoed", logging.RequestIDHeader, got)
	}
}

func TestRouterReplacesUnusableRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	// Control characters would corrupt log lines, so the ID must be regenerated.
	req.Header.Set(logging.RequestIDHeader, "bad\nid")

	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, req)

	got := rec.Header().Get(logging.RequestIDHeader)
	if got == "bad\nid" || len(got) != 32 {
		t.Fatalf("%s = %q, want a regenerated id", logging.RequestIDHeader, got)
	}
}

func TestEchoRejectsOversizedBody(t *testing.T) {
	registry, err := metrics.NewRegistry(metrics.WithNamespace("test"))
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	// A 16-byte cap is smaller than the payload below.
	router := httptransport.Router(echo.NewService(), registry, 16)

	body := `{"message":"` + strings.Repeat("A", 512) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestEchoRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"message":"hi","evil":1}`))
	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if strings.Contains(rec.Body.String(), "evil") {
		t.Fatalf("body %q leaks decoder detail", rec.Body.String())
	}
}

func TestEchoRoundTrip(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"message":" hi "}`))
	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body %q", rec.Code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"hi"}` {
		t.Fatalf("body = %s", got)
	}
}
