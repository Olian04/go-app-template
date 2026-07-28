// [[ when (modeIs "http") ]]
package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"[[.ModulePath]]/internal/observability/logging"
	"[[.ModulePath]]/internal/observability/metrics"
)

// Middleware wraps a handler. Compose with chain.
type Middleware func(stdhttp.Handler) stdhttp.Handler

// chain applies middlewares so the first listed is the outermost wrapper: the
// request travels through them in argument order.
func chain(h stdhttp.Handler, mws ...Middleware) stdhttp.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// recordingWriter captures status and byte count for logging and metrics.
//
// WriteHeader may never be called, so status defaults to 200 to match what
// net/http actually sends.
type recordingWriter struct {
	stdhttp.ResponseWriter
	status int
	bytes  int
}

func (w *recordingWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *recordingWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = stdhttp.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (w *recordingWriter) statusCode() int {
	if w.status == 0 {
		return stdhttp.StatusOK
	}
	return w.status
}

// Recover turns a handler panic into a 500 instead of dropping the connection,
// so one bad request cannot take the server down.
func Recover() Middleware {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			defer func() {
				if v := recover(); v != nil {
					logging.FromContext(r.Context()).Error("http panic recovered",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.Any("panic", v),
					)
					stdhttp.Error(w, "internal server error", stdhttp.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID attaches a correlation ID to the request context, response header,
// and every log line emitted downstream. An inbound X-Request-ID is trusted so
// IDs survive across service hops; otherwise one is generated.
func RequestID() Middleware {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			id := logging.SanitizeRequestID(r.Header.Get(logging.RequestIDHeader))
			if id == "" {
				id = logging.NewRequestID()
			}
			ctx := logging.WithRequestID(r.Context(), id)
			w.Header().Set(logging.RequestIDHeader, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logging emits one structured line per request, including the correlation ID.
func Logging() Middleware {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			rec := &recordingWriter{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			logging.FromContext(r.Context()).LogAttrs(r.Context(), slog.LevelInfo, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.statusCode()),
				slog.Int("bytes", rec.bytes),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

// Metrics records request count, in-flight gauge, and latency histogram.
func Metrics(registry *metrics.Registry) Middleware {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			registry.IncInFlight()
			defer registry.DecInFlight()

			rec := &recordingWriter{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			registry.ObserveRequest(r.Method, rec.statusCode(), time.Since(start))
		})
	}
}
