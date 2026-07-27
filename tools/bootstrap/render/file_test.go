package render_test

import (
	"strings"
	"testing"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

func TestFile_whenModeIs_match(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"cli\" \"cli-library\") ]]\ncfg: [[ .ModulePath ]]\n")
	ctx := render.Context{
		Mode:       render.ModeCLI,
		ModulePath: "example.com/app",
	}
	out, skipped, err := render.File("cfg.yaml", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	got := string(out)
	if strings.Contains(got, "when (modeIs") || strings.Contains(got, "whenMode") {
		t.Fatalf("gate line not stripped: %q", got)
	}
	if got != "cfg: example.com/app\n" {
		t.Fatalf("got %q", got)
	}
}

func TestFile_whenModeIs_noMatch(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"http\") ]]\ncfg: x\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("cfg.yaml", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if !skipped {
		t.Fatal("expected skipped")
	}
	if out != nil {
		t.Fatalf("expected nil out, got %q", out)
	}
}

func TestFile_whenModeIs_multiOR(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"cli\" \"http\") ]]\nbody [[ .CliName ]]\n")
	ctx := render.Context{
		Mode:    render.ModeHTTP,
		CliName: "mycli",
	}
	out, skipped, err := render.File("x.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included via OR modes")
	}
	if string(out) != "body mycli\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_whenModeIs_allFalse(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"cli\" \"http\") ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeLibrary}
	_, skipped, err := render.File("x.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if !skipped {
		t.Fatal("expected skipped")
	}
}

func TestFile_whenModeIs_single(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"cli\") ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("x.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "body\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_gateIsolation_rejectsModulePath(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when .ModulePath ]]\nbody\n")
	ctx := render.Context{
		Mode:       render.ModeCLI,
		ModulePath: "example.com/app",
	}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected error: gate must not see ModulePath")
	}
}

func TestFile_stripGateLine(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when (modeIs \"cli\") ]]\nline2\nline3\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("x.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "line2\nline3\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_ungated_noStrip(t *testing.T) {
	t.Parallel()
	src := []byte("package main\nmodule [[ .ModulePath ]]\n")
	ctx := render.Context{Mode: render.ModeCLI, ModulePath: "example.com/app"}
	out, skipped, err := render.File("main.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("ungated must include")
	}
	if string(out) != "package main\nmodule example.com/app\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_normalRejectsWhen(t *testing.T) {
	t.Parallel()
	src := []byte("package main\n[[ when (modeIs \"cli\") ]]\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("main.go", src, ctx)
	if err == nil {
		t.Fatal("expected error: normal mode must reject when")
	}
}

func TestFile_normalRejectsWhenEither(t *testing.T) {
	t.Parallel()
	src := []byte("package main\n[[ whenEither true false ]]\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("main.go", src, ctx)
	if err == nil {
		t.Fatal("expected error: normal mode must reject whenEither")
	}
}

func TestFile_normalIfModeIs(t *testing.T) {
	t.Parallel()
	src := []byte("hub\n[[ if modeIs \"cli\" ]]cli[[ end ]][[ if modeIs \"http\" ]]http[[ end ]]\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("hub.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "hub\ncli\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_invalidGateHardError(t *testing.T) {
	t.Parallel()
	src := []byte("[[ .Mode ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected hard error for gate without when")
	}
}

func TestFile_whenMode_rejected(t *testing.T) {
	t.Parallel()
	src := []byte("[[ whenMode \"cli\" ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected hard error for removed whenMode")
	}
}

func TestFile_whenEither_rejected(t *testing.T) {
	t.Parallel()
	src := []byte("[[ whenEither true false ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected hard error for removed whenEither")
	}
}

func TestFile_binaryAllowlist_passthrough(t *testing.T) {
	t.Parallel()
	src := []byte{0x89, 0x50, 0x4e, 0x47, 0x0a, '[', '['}
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("logo.png", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("binary must not skip")
	}
	if string(out) != string(src) {
		t.Fatalf("binary mutated: %v", out)
	}
}

func TestFile_commentPrefixedGate_hash(t *testing.T) {
	t.Parallel()
	src := []byte("# [[ when (modeIs \"cli\" \"library\" \"cli-library\" \"http\") ]]\nname: CI\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("ci.yml", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "name: CI\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_commentPrefixedGate_slash(t *testing.T) {
	t.Parallel()
	src := []byte("// [[ when (modeIs \"http\") ]]\npackage http\n")
	ctx := render.Context{Mode: render.ModeHTTP}
	out, skipped, err := render.File("router.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "package http\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_commentPrefixedGate_block(t *testing.T) {
	t.Parallel()
	src := []byte("/* [[ when (modeIs \"library\" \"cli-library\") ]] */\npackage lib\n")
	ctx := render.Context{Mode: render.ModeLibrary}
	out, skipped, err := render.File("greet.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "package lib\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_commentPrefixedGate_html(t *testing.T) {
	t.Parallel()
	src := []byte("<!-- [[ when (modeIs \"cli\" \"library\" \"cli-library\" \"http\") ]] -->\n# Agents\n")
	ctx := render.Context{Mode: render.ModeHTTP}
	out, skipped, err := render.File("AGENTS.md", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "# Agents\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_commentPrefixedGate_falseSkips(t *testing.T) {
	t.Parallel()
	src := []byte("# [[ when (modeIs \"http\") ]]\nname: CI\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, skipped, err := render.File("ci.yml", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if !skipped {
		t.Fatal("expected skipped")
	}
}

func TestFile_commentPrefixed_whenModeIs(t *testing.T) {
	t.Parallel()
	src := []byte("# [[ when (modeIs \"cli\" \"http\") ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("x.yml", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "body\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_multipleGatesHardError(t *testing.T) {
	t.Parallel()
	src := []byte("# [[ when (modeIs \"cli\") ]] [[ when (modeIs \"http\") ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected hard error for multiple gate actions")
	}
}

func TestFile_invalidGateAttempt_hyphen(t *testing.T) {
	t.Parallel()
	src := []byte("# [[ when-mode \"cli\" \"http\" ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, _, err := render.File("x.yml", src, ctx)
	if err == nil {
		t.Fatal("expected hard error for when-mode")
	}
}

func TestFile_ungated_normalActionOnLine1(t *testing.T) {
	t.Parallel()
	src := []byte("module [[ .ModulePath ]]\n\ngo 1.26\n")
	ctx := render.Context{Mode: render.ModeCLI, ModulePath: "example.com/app"}
	out, skipped, err := render.File("go.mod", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("ungated must include")
	}
	if string(out) != "module example.com/app\n\ngo 1.26\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_loggingGate_include(t *testing.T) {
	t.Parallel()
	src := []byte("// [[ when (modeIs \"cli\" \"cli-library\" \"http\") ]]\npackage logging\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("logging.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "package logging\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_loggingGate_skip(t *testing.T) {
	t.Parallel()
	src := []byte("// [[ when (modeIs \"cli\" \"cli-library\" \"http\") ]]\npackage logging\n")
	ctx := render.Context{Mode: render.ModeLibrary}
	_, skipped, err := render.File("logging.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if !skipped {
		t.Fatal("expected skipped")
	}
}

func TestFile_metricsGate_include(t *testing.T) {
	t.Parallel()
	src := []byte("// [[ when (modeIs \"http\") ]]\npackage metrics\n")
	ctx := render.Context{Mode: render.ModeHTTP}
	out, skipped, err := render.File("registry.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "package metrics\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_metricsGate_skip(t *testing.T) {
	t.Parallel()
	src := []byte("// [[ when (modeIs \"http\") ]]\npackage metrics\n")
	ctx := render.Context{Mode: render.ModeCLI}
	_, skipped, err := render.File("registry.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if !skipped {
		t.Fatal("expected skipped")
	}
}

func TestFile_bodyIf_modeIs(t *testing.T) {
	t.Parallel()
	src := []byte("hub\n[[ if modeIs \"cli\" \"cli-library\" \"http\" ]]log[[ end ]][[ if modeIs \"http\" ]]met[[ end ]]\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("hub.txt", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "hub\nlog\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_when_bool(t *testing.T) {
	t.Parallel()
	src := []byte("[[ when true ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeCLI}
	out, skipped, err := render.File("x.go", src, ctx)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if skipped {
		t.Fatal("expected included")
	}
	if string(out) != "body\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFile_bareWhenModeIs_noRewrite(t *testing.T) {
	t.Parallel()
	// Bare when modeIs … is invalid Go template nesting; no paren rewriter.
	src := []byte("[[ when modeIs \"http\" ]]\nbody\n")
	ctx := render.Context{Mode: render.ModeHTTP}
	_, _, err := render.File("x.go", src, ctx)
	if err == nil {
		t.Fatal("expected error for bare when modeIs without parentheses")
	}
}
