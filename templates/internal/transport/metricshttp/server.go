// [[ when (modeIs "http") ]]
// Package metricshttp exposes the Prometheus scrape endpoint as a handler.
//
// It deliberately owns no server lifecycle: the composition root (internal/app)
// runs every listener through one code path, so timeouts and shutdown behaviour
// cannot drift between the service and metrics listeners.
package metricshttp

import "net/http"

// Handler serves the registry at /metrics and redirects / there for convenience.
func Handler(metrics http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusFound)
	}))
	mux.Handle("GET /metrics", metrics)
	return mux
}
