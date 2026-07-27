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

func TestBuildContext_binariesByMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		sel     selection
		wantLen int
		wantBin string
	}{
		{
			name: "cli",
			sel: selection{
				Mode: render.ModeCLI, ModulePath: "example.com/app", GoVersion: "1.26", CliName: "app",
			},
			wantLen: 1, wantBin: "app",
		},
		{
			name: "library",
			sel: selection{
				Mode: render.ModeLibrary, ModulePath: "example.com/app", GoVersion: "1.26", LibName: "app",
			},
			wantLen: 0,
		},
		{
			name: "cli-library",
			sel: selection{
				Mode: render.ModeCLILibrary, ModulePath: "example.com/app", GoVersion: "1.26",
				CliName: "app", LibName: "app",
			},
			wantLen: 1, wantBin: "app",
		},
		{
			name: "http",
			sel: selection{
				Mode: render.ModeHTTP, ModulePath: "example.com/app", GoVersion: "1.26", ServiceName: "appsvc",
			},
			wantLen: 1, wantBin: "appsvc",
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
			if len(ctx.Binaries) != tc.wantLen {
				t.Fatalf("Binaries len=%d, want %d (%+v)", len(ctx.Binaries), tc.wantLen, ctx.Binaries)
			}
			if tc.wantLen == 1 && ctx.Binaries[0].Name != tc.wantBin {
				t.Fatalf("Binary.Name=%q, want %q", ctx.Binaries[0].Name, tc.wantBin)
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
