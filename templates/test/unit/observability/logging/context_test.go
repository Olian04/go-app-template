// [[ when (modeIs "cli" "cli-library" "http") ]]
package logging_test

import (
	"context"
	"testing"

	"[[.ModulePath]]/internal/observability/logging"
)

func TestRequestIDRoundTrip(t *testing.T) {
	ctx := logging.WithRequestID(context.Background(), "abc123")
	if got := logging.RequestIDFromContext(ctx); got != "abc123" {
		t.Fatalf("got %q, want abc123", got)
	}
}

func TestRequestIDMissingIsEmpty(t *testing.T) {
	if got := logging.RequestIDFromContext(context.Background()); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestNewRequestIDIsUniqueHex(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for range 100 {
		id := logging.NewRequestID()
		if len(id) != 32 {
			t.Fatalf("id %q length = %d, want 32", id, len(id))
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestSanitizeRequestID(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"plain":            {"trace-42", "trace-42"},
		"empty":            {"", ""},
		"newline":          {"bad\nid", ""},
		"tab":              {"bad\tid", ""},
		"space":            {"bad id", ""},
		"non-ascii":        {"trace-π", ""},
		"too long":         {string(make([]byte, 200)), ""},
		"printable symbol": {"a:b/c=1", "a:b/c=1"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := logging.SanitizeRequestID(tc.in); got != tc.want {
				t.Fatalf("SanitizeRequestID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// FromContext must never return nil, so callers can log unconditionally.
func TestFromContextAlwaysReturnsLogger(t *testing.T) {
	if logging.FromContext(context.Background()) == nil {
		t.Fatal("nil logger without request id")
	}
	if logging.FromContext(logging.WithRequestID(context.Background(), "x")) == nil {
		t.Fatal("nil logger with request id")
	}
}
