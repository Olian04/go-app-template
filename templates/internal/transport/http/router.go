// [[ when (modeIs "http") ]]
package http

import (
	stdhttp "net/http"

	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/metrics"
	"[[.ModulePath]]/internal/transport/http/handlers"
)

// Router builds the service handler. maxBodyBytes bounds decoded request bodies.
func Router(svc echo.Service, registry *metrics.Registry, maxBodyBytes int64) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	mux.Handle("POST /echo", handlers.NewEchoHandler(svc, maxBodyBytes))

	// Recover outermost so it also catches panics raised in the layers below;
	// RequestID ahead of Logging/Metrics so both see the correlation ID.
	return chain(mux,
		Recover(),
		RequestID(),
		Logging(),
		Metrics(registry),
	)
}
