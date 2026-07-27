// [[ when (modeIs "library" "cli-library") ]]
package [[.LibName]]_test

import (
	"testing"

	"[[.ModulePath]]/pkg/[[.LibName]]"
)

func TestGreet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "hello"},
		{name: "named", in: "world", want: "hello, world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := [[.LibName]].Greet(tt.in)
			if got != tt.want {
				t.Fatalf("Greet(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
