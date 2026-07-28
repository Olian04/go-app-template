// [[ when (modeIs "cli" "cli-library" "http") ]]
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

// RequestIDHeader carries the correlation ID between services.
const RequestIDHeader = "X-Request-ID"

// requestIDKey is unexported so only this package can seed the value.
type requestIDKey struct{}

// maxRequestIDLen bounds an inbound ID so a caller cannot bloat every log line.
const maxRequestIDLen = 128

// NewRequestID returns a random 128-bit hex ID.
func NewRequestID() string {
	var b [16]byte
	// rand.Read never returns an error (it panics internally on failure).
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// SanitizeRequestID keeps an inbound ID only if it is safe to echo back and log:
// printable ASCII, no whitespace or control characters, length-bounded. Returns
// "" when the value is unusable, signalling the caller to generate a fresh ID.
func SanitizeRequestID(s string) string {
	if s == "" || len(s) > maxRequestIDLen {
		return ""
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; c < 0x21 || c > 0x7e {
			return ""
		}
	}
	return s
}

// WithRequestID stores id on ctx for FromContext to pick up.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the correlation ID, or "" when unset.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// FromContext returns the default logger annotated with the context's
// correlation ID when one is present. Use it instead of slog's package-level
// functions so request-scoped lines stay correlated.
func FromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if id := RequestIDFromContext(ctx); id != "" {
		return logger.With(slog.String("request_id", id))
	}
	return logger
}
