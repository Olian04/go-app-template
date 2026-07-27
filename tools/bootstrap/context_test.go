package main

import (
	"testing"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

func TestParseMode(t *testing.T) {
	t.Parallel()
	for _, want := range []render.Mode{
		render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP,
	} {
		got, err := parseMode(string(want))
		if err != nil {
			t.Fatalf("parseMode(%q): %v", want, err)
		}
		if got != want {
			t.Fatalf("parseMode(%q)=%q, want %q", want, got, want)
		}
	}
	if _, err := parseMode("nope"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestBuildContext_binaryByMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		sel     selection
		wantBin string // empty => Binary nil
	}{
		{
			name: "cli",
			sel: selection{
				Mode: render.ModeCLI, ModulePath: "example.com/app", GoVersion: "1.26", CliName: "app",
			},
			wantBin: "app",
		},
		{
			name: "library",
			sel: selection{
				Mode: render.ModeLibrary, ModulePath: "example.com/app", GoVersion: "1.26", LibName: "app",
			},
		},
		{
			name: "cli-library",
			sel: selection{
				Mode: render.ModeCLILibrary, ModulePath: "example.com/app", GoVersion: "1.26",
				CliName: "app", LibName: "app",
			},
			wantBin: "app",
		},
		{
			name: "http",
			sel: selection{
				Mode: render.ModeHTTP, ModulePath: "example.com/app", GoVersion: "1.26", ServiceName: "appsvc",
			},
			wantBin: "appsvc",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, err := buildContext(tc.sel)
			if err != nil {
				t.Fatalf("buildContext: %v", err)
			}
			if ctx.Mode != tc.sel.Mode {
				t.Fatalf("Mode=%q, want %q", ctx.Mode, tc.sel.Mode)
			}
			if tc.wantBin == "" {
				if ctx.Binary != nil {
					t.Fatalf("Binary=%+v, want nil", ctx.Binary)
				}
				return
			}
			if ctx.Binary == nil {
				t.Fatal("Binary=nil, want set")
			}
			if ctx.Binary.Name != tc.wantBin {
				t.Fatalf("Binary.Name=%q, want %q", ctx.Binary.Name, tc.wantBin)
			}
		})
	}
}

func TestApplyNameDefaults_fromBasename(t *testing.T) {
	t.Parallel()
	sel := selection{Mode: render.ModeCLILibrary, ModulePath: "github.com/acme/Cool-App"}
	applyNameDefaults(&sel)
	if sel.CliName != "cool-app" || sel.LibName != "cool-app" {
		t.Fatalf("defaults: cli=%q lib=%q, want cool-app", sel.CliName, sel.LibName)
	}
}

func TestValidateContext_httpOK(t *testing.T) {
	t.Parallel()
	ctx := render.Context{
		Mode:        render.ModeHTTP,
		ModulePath:  "example.com/smoke/app",
		GoVersion:   "1.26",
		ServiceName: "smokeapp",
	}
	if err := validateContext(ctx); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
