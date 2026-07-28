package echo_test

import (
	"encoding/json"
	"testing"

	"[[.ModulePath]]/internal/domain/echo"
)

func TestEcho(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "trims surrounding space", in: " hello ", want: "hello"},
		{name: "leaves inner space", in: "hello world", want: "hello world"},
		{name: "empty stays empty", in: "", want: ""},
		{name: "whitespace only", in: "   ", want: ""},
	}

	svc := echo.NewService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := svc.Echo(echo.Request{Message: tt.in}).Message; got != tt.want {
				t.Fatalf("Echo(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// The JSON tags are part of the domain contract in every mode, so a transport
// can decode straight into Request without an intermediate type.
func TestEchoJSONRoundTrip(t *testing.T) {
	t.Parallel()

	var req echo.Request
	if err := json.Unmarshal([]byte(`{"message":" hello "}`), &req); err != nil {
		t.Fatal(err)
	}

	out, err := json.Marshal(echo.NewService().Echo(req))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"message":"hello"}` {
		t.Fatalf("marshalled %s", out)
	}
}
