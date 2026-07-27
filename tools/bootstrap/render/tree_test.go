package render_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

func TestTree_pathExpansion(t *testing.T) {
	t.Parallel()
	in := t.TempDir()
	out := t.TempDir()
	srcDir := filepath.Join(in, "cmd", "[[.CliName]]")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := render.Context{
		Mode:    render.ModeCLI,
		CliName: "toolbox",
	}
	if err := render.Tree(in, out, ctx); err != nil {
		t.Fatalf("Tree: %v", err)
	}
	got := filepath.Join(out, "cmd", "toolbox", "main.go")
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("missing expanded path %s: %v", got, err)
	}
}

func TestTree_duplicateOutPath(t *testing.T) {
	t.Parallel()
	in := t.TempDir()
	out := t.TempDir()
	a := filepath.Join(in, "[[.CliName]].go")
	b := filepath.Join(in, "same.go")
	if err := os.WriteFile(a, []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("package b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := render.Context{
		Mode:    render.ModeCLI,
		CliName: "same",
	}
	err := render.Tree(in, out, ctx)
	if err == nil {
		t.Fatal("expected duplicate out path error")
	}
}

func TestTree_emptyDirPrune(t *testing.T) {
	t.Parallel()
	in := t.TempDir()
	out := t.TempDir()
	empty := filepath.Join(in, "cmd", "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(in, "pkg")
	if err := os.MkdirAll(keep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(keep, "lib.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// gated file under cmd that will be skipped → cmd should not appear
	gated := filepath.Join(in, "cmd", "cli", "main.go")
	if err := os.MkdirAll(filepath.Dir(gated), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gated, []byte("[[ when (modeIs \"cli\") ]]\npackage main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := render.Context{Mode: render.ModeLibrary}
	if err := render.Tree(in, out, ctx); err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "pkg", "lib.go")); err != nil {
		t.Fatalf("kept file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "cmd")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd pruned, err=%v", err)
	}
}

func TestTree_skipsGatedAndWritesOthers(t *testing.T) {
	t.Parallel()
	in := t.TempDir()
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(in, "always.go"), []byte("package always\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(in, "cli.go"), []byte("[[ when (modeIs \"cli\") ]]\npackage cli\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := render.Context{Mode: render.ModeLibrary}
	if err := render.Tree(in, out, ctx); err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "always.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "cli.go")); !os.IsNotExist(err) {
		t.Fatal("cli.go should be skipped")
	}
}

func TestTree_binaryCopied(t *testing.T) {
	t.Parallel()
	in := t.TempDir()
	out := t.TempDir()
	raw := []byte{0x00, 0x01, 0x02, '[', '['}
	if err := os.WriteFile(filepath.Join(in, "icon.ico"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := render.Context{Mode: render.ModeCLI}
	if err := render.Tree(in, out, ctx); err != nil {
		t.Fatalf("Tree: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, "icon.ico"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatalf("binary changed: %v", got)
	}
}
