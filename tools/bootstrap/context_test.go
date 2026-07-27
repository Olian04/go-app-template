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
	if sel.CliName != "cool-app" {
		t.Fatalf("CliName=%q, want cool-app", sel.CliName)
	}
	if sel.LibName != "coolapp" {
		t.Fatalf("LibName=%q, want coolapp (Go package name)", sel.LibName)
	}
}

func TestGoPackageName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"go-app-template", "goapptemplate"},
		{"cool-app", "coolapp"},
		{"alreadyok", "alreadyok"},
		{"", "lib"},
		{"123bad", "lib123bad"},
		{"foo_bar", "foobar"},
	}
	for _, tc := range cases {
		if got := goPackageName(tc.in); got != tc.want {
			t.Fatalf("goPackageName(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidatePackageName_rejectsHyphen(t *testing.T) {
	t.Parallel()
	if err := validatePackageName("lib-name", "go-app-template"); err == nil {
		t.Fatal("expected error for hyphenated lib-name")
	}
	if err := validatePackageName("lib-name", "goapptemplate"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestPrometheusMetricPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"go-app-template", "go_app_template"},
		{"cool-app", "cool_app"},
		{"already_ok", "already_ok"},
		{"", "app"},
		{"123bad", "app_123bad"},
		{"foo.bar", "foo_bar"},
	}
	for _, tc := range cases {
		if got := prometheusMetricPrefix(tc.in); got != tc.want {
			t.Fatalf("prometheusMetricPrefix(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildContext_metricPrefix(t *testing.T) {
	t.Parallel()
	ctx, err := buildContext(selection{
		Mode: render.ModeHTTP, ModulePath: "github.com/Olian04/go-app-template",
		GoVersion: "1.26", ServiceName: "go-app-template",
	})
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if ctx.ModuleBasename != "go-app-template" {
		t.Fatalf("ModuleBasename=%q", ctx.ModuleBasename)
	}
	if ctx.MetricPrefix != "go_app_template" {
		t.Fatalf("MetricPrefix=%q, want go_app_template", ctx.MetricPrefix)
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
