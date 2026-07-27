// [[ when (modeIs "http") ]]
package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"[[.ModulePath]]/internal/domain/echo"
[[ if modeIs "http" ]]
	"[[.ModulePath]]/internal/observability/metrics"
[[ end ]]
	"[[.ModulePath]]/internal/transport/http/handlers"
)

func Router(svc echo.Service[[ if modeIs "http" ]], registry *metrics.Registry[[ end ]]) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	echoHandler := handlers.NewEchoHandler(svc)
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	mux.Handle("POST /echo", echoHandler)
	return middleware(mux[[ if modeIs "http" ]], registry[[ end ]])
}

func middleware(next stdhttp.Handler[[ if modeIs "http" ]], registry *metrics.Registry[[ end ]]) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()
[[ if modeIs "http" ]]
		registry.IncRequests()
[[ end ]]
		next.ServeHTTP(w, r)
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
