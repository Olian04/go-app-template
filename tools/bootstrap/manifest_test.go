package main

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

// updateManifests rewrites the golden files instead of asserting against them:
//
//	go test . -run TestModeManifest -update
var updateManifests = flag.Bool("update", false, "rewrite mode manifest golden files")

// manifestSelection is the fixed input for every mode, so a manifest diff can
// only come from a gate change and never from differing names.
func manifestSelection(mode render.Mode) selection {
	return selection{
		Mode:        mode,
		ModulePath:  "example.com/manifest/app",
		GoVersion:   "1.26",
		CliName:     "appcli",
		ServiceName: "appsvc",
		LibName:     "applib",
	}
}

// TestModeManifest pins the exact file set each mode renders.
//
// Mode gates are the template's only feature switch, so a mis-gated file (one
// that leaks into a mode that cannot use it, or goes missing from one that
// needs it) is a silent correctness bug: the generated project still compiles.
// Comparing whole manifests makes that class of change fail loudly and forces
// it to be reviewed as a deliberate diff.
func TestModeManifest(t *testing.T) {
	templatesDir := repoTemplatesDir(t)

	for _, mode := range []render.Mode{
		render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP,
	} {
		t.Run(string(mode), func(t *testing.T) {
			ctx, err := buildContext(manifestSelection(mode))
			if err != nil {
				t.Fatalf("buildContext: %v", err)
			}
			if err := validateContext(ctx); err != nil {
				t.Fatalf("validateContext: %v", err)
			}

			outDir := t.TempDir()
			if err := render.Tree(templatesDir, outDir, ctx); err != nil {
				t.Fatalf("render.Tree: %v", err)
			}

			got := renderedPaths(t, outDir)
			goldenPath := filepath.Join("testdata", "manifest", string(mode)+".txt")

			if *updateManifests {
				writeManifest(t, goldenPath, got)
				return
			}

			want := readManifest(t, goldenPath)
			assertSameManifest(t, mode, want, got)
		})
	}
}

// TestModeManifestGatingInvariants states the mode/feature rules in prose form.
// The golden manifests catch any change; these cases say which changes are
// actually wrong, so a failure explains itself without diffing a file list.
func TestModeManifestGatingInvariants(t *testing.T) {
	templatesDir := repoTemplatesDir(t)

	rules := []struct {
		path    string
		reason  string
		present []render.Mode
	}{
		{
			path:    "internal/config/http.go",
			reason:  "HTTP tuning only applies where a server runs",
			present: []render.Mode{render.ModeHTTP},
		},
		{
			path:    "internal/config/metrics.go",
			reason:  "metrics are scraped from the HTTP service only",
			present: []render.Mode{render.ModeHTTP},
		},
		{
			path:    "internal/transport/http/middleware.go",
			reason:  "middleware chain belongs to the HTTP transport",
			present: []render.Mode{render.ModeHTTP},
		},
		{
			path:    "internal/app/bootstrap.go",
			reason:  "composition root exists only for the service",
			present: []render.Mode{render.ModeHTTP},
		},
		{
			path:    "internal/config/config.go",
			reason:  "config backs every mode that produces a binary",
			present: []render.Mode{render.ModeCLI, render.ModeCLILibrary, render.ModeHTTP},
		},
		{
			// The domain is the fixed point every mode adapts to; only the IO
			// around it is mode-specific.
			path:    "internal/domain/echo/service.go",
			reason:  "the demo domain model ships in every mode",
			present: []render.Mode{render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP},
		},
		{
			path:    "test/unit/domain/echo/service_test.go",
			reason:  "domain tests ship wherever the domain does",
			present: []render.Mode{render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP},
		},
		{
			path:    "internal/observability/logging/context.go",
			reason:  "correlation-id helpers ship with logging in every binary mode",
			present: []render.Mode{render.ModeCLI, render.ModeCLILibrary, render.ModeHTTP},
		},
		{
			path:    "cmd/appcli/main.go",
			reason:  "CLI entrypoint belongs to the cli modes",
			present: []render.Mode{render.ModeCLI, render.ModeCLILibrary},
		},
		{
			path:    "cmd/appsvc/main.go",
			reason:  "service entrypoint belongs to http",
			present: []render.Mode{render.ModeHTTP},
		},
		{
			path:    "pkg/applib/echo.go",
			reason:  "the public facade over the domain belongs to the library modes",
			present: []render.Mode{render.ModeLibrary, render.ModeCLILibrary},
		},
		{
			path:    ".github/workflows/ci.yml",
			reason:  "CI is ungated and ships in every mode",
			present: []render.Mode{render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP},
		},
	}

	for _, mode := range []render.Mode{
		render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP,
	} {
		ctx, err := buildContext(manifestSelection(mode))
		if err != nil {
			t.Fatalf("buildContext(%s): %v", mode, err)
		}
		outDir := t.TempDir()
		if err := render.Tree(templatesDir, outDir, ctx); err != nil {
			t.Fatalf("render.Tree(%s): %v", mode, err)
		}
		rendered := renderedPaths(t, outDir)

		for _, rule := range rules {
			wantPresent := slices.Contains(rule.present, mode)
			gotPresent := slices.Contains(rendered, rule.path)
			if wantPresent == gotPresent {
				continue
			}
			if wantPresent {
				t.Errorf("mode %s: %s missing (%s)", mode, rule.path, rule.reason)
			} else {
				t.Errorf("mode %s: %s should not render (%s)", mode, rule.path, rule.reason)
			}
		}
	}
}

// TestModeManifestNoLeftoverGates guards the dialect itself: an unresolved
// delimiter in rendered output means a gate or action silently failed to apply.
func TestModeManifestNoLeftoverGates(t *testing.T) {
	templatesDir := repoTemplatesDir(t)

	for _, mode := range []render.Mode{
		render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP,
	} {
		t.Run(string(mode), func(t *testing.T) {
			ctx, err := buildContext(manifestSelection(mode))
			if err != nil {
				t.Fatalf("buildContext: %v", err)
			}
			outDir := t.TempDir()
			if err := render.Tree(templatesDir, outDir, ctx); err != nil {
				t.Fatalf("render.Tree: %v", err)
			}

			for _, rel := range renderedPaths(t, outDir) {
				data, err := os.ReadFile(filepath.Join(outDir, rel))
				if err != nil {
					t.Fatalf("read %s: %v", rel, err)
				}
				// The awk snippet in the Makefile help target is the one place a
				// literal "[[" legitimately survives rendering.
				if rel == "Makefile" {
					continue
				}
				if strings.Contains(string(data), "[[") || strings.Contains(string(data), "]]") {
					t.Errorf("%s: rendered output still contains template delimiters", rel)
				}
			}
		})
	}
}

func repoTemplatesDir(t *testing.T) string {
	t.Helper()
	// tools/bootstrap -> repo root
	dir, err := filepath.Abs(filepath.Join("..", "..", "templates"))
	if err != nil {
		t.Fatalf("templates dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("templates dir %s: %v", dir, err)
	}
	return dir
}

// renderedPaths lists rendered files as sorted slash-separated relative paths.
func renderedPaths(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	slices.Sort(out)
	return out
}

func readManifest(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v (run: go test . -run TestModeManifest -update)", path, err)
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	slices.Sort(out)
	return out
}

func writeManifest(t *testing.T, path string, paths []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	body := "# Generated by: go test . -run TestModeManifest -update\n" +
		"# Files this mode renders. A diff here is a mode-enablement change.\n" +
		strings.Join(paths, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	t.Logf("updated %s (%d paths)", path, len(paths))
}

func assertSameManifest(t *testing.T, mode render.Mode, want, got []string) {
	t.Helper()
	for _, p := range got {
		if !slices.Contains(want, p) {
			t.Errorf("mode %s: unexpected rendered file %q", mode, p)
		}
	}
	for _, p := range want {
		if !slices.Contains(got, p) {
			t.Errorf("mode %s: expected file %q was not rendered", mode, p)
		}
	}
	if t.Failed() {
		t.Logf("re-record with: go test . -run TestModeManifest -update")
	}
}
