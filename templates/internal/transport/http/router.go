// [[ when (modeIs "http") ]]
package http

import (
	stdhttp "net/http"

	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/metrics"
	"[[.ModulePath]]/internal/transport/http/handlers"
)

// Options are the transport's knobs. Plain values rather than a config type, so
// the transport layer stays independent of how configuration is loaded.
type Options struct {
	// MaxBodyBytes bounds decoded request bodies.
	MaxBodyBytes int64
	// DocsEnabled serves /docs and the OpenAPI spec. Turn it off to keep the
	// contract private in an untrusted network.
	DocsEnabled bool
}

// Router builds the service handler.
//
// Returns an error when documentation is enabled but the embedded OpenAPI spec
// does not parse, so a broken contract fails at startup rather than on request.
func Router(svc echo.Service, registry *metrics.Registry, opts Options) (stdhttp.Handler, error) {
	mux := stdhttp.NewServeMux()
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	mux.Handle("POST /echo", handlers.NewEchoHandler(svc, opts.MaxBodyBytes))

	if opts.DocsEnabled {
		d, err := newDocs()
		if err != nil {
			return nil, err
		}
		d.register(mux)
	}

	// Recover outermost so it also catches panics raised in the layers below;
	// RequestID ahead of Logging/Metrics so both see the correlation ID.
	return chain(mux,
		Recover(),
		RequestID(),
		Logging(),
		Metrics(registry),
	), nil
}
