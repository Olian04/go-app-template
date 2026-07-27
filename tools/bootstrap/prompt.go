package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
	"github.com/charmbracelet/gum/choose"
	"github.com/charmbracelet/gum/input"
	gumstyle "github.com/charmbracelet/gum/style"
)

var modeChoices = []string{
	string(render.ModeCLI),
	string(render.ModeLibrary),
	string(render.ModeCLILibrary),
	string(render.ModeHTTP),
}

func promptInteractive(base selection) (selection, error) {
	sel := base

	if sel.Mode == "" {
		picked, err := gumChoose("Select mode", modeChoices)
		if err != nil {
			return sel, fmt.Errorf("mode select: %w", err)
		}
		mode, err := parseMode(picked)
		if err != nil {
			return sel, err
		}
		sel.Mode = mode
	}

	defaultMod := sel.ModulePath
	if defaultMod == "" {
		if hint, err := defaultModulePath(); err == nil {
			defaultMod = hint
		}
	}
	mod, err := gumInput("Go module path", "e.g. github.com/you/your-app", defaultMod)
	if err != nil {
		return sel, fmt.Errorf("module path: %w", err)
	}
	sel.ModulePath = strings.TrimSpace(mod)
	baseName := moduleBasename(sel.ModulePath)

	if modeNeedsHTTP(sel.Mode) {
		def := sel.ServiceName
		if def == "" {
			def = baseName
		}
		name, err := gumInput("HTTP service name (cmd/ directory)", "e.g. myapi", def)
		if err != nil {
			return sel, fmt.Errorf("service name: %w", err)
		}
		sel.ServiceName = strings.TrimSpace(name)
	}
	if modeNeedsCLI(sel.Mode) {
		def := sel.CliName
		if def == "" {
			def = baseName
		}
		name, err := gumInput("CLI name (cmd/ directory)", "e.g. mycli", def)
		if err != nil {
			return sel, fmt.Errorf("cli name: %w", err)
		}
		sel.CliName = strings.TrimSpace(name)
	}
	if modeNeedsLibrary(sel.Mode) {
		def := sel.LibName
		if def == "" {
			def = goPackageName(baseName)
		}
		name, err := gumInput("Library Go package name", "e.g. mylib (no hyphens)", def)
		if err != nil {
			return sel, fmt.Errorf("lib name: %w", err)
		}
		sel.LibName = strings.TrimSpace(name)
	}

	return sel, nil
}

func gumChoose(header string, options []string) (string, error) {
	opts := choose.Options{
		Options:           options,
		Limit:             1,
		Header:            header,
		Height:            len(options) + 2,
		ShowHelp:          true,
		Padding:           "0 0",
		Cursor:            "> ",
		CursorPrefix:      "• ",
		SelectedPrefix:    "✓ ",
		UnselectedPrefix:  "• ",
		OutputDelimiter:   "\n",
		InputDelimiter:    "\n",
		CursorStyle:       gumstyle.Styles{Foreground: "212"},
		HeaderStyle:       gumstyle.Styles{Foreground: "99"},
		SelectedItemStyle: gumstyle.Styles{Foreground: "212"},
	}
	out, err := captureStdout(opts.Run)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func gumInput(header, placeholder, value string) (string, error) {
	opts := input.Options{
		Header:           header,
		Placeholder:      placeholder,
		Value:            value,
		Prompt:           "> ",
		ShowHelp:         true,
		Padding:          "0 0",
		CharLimit:        400,
		CursorMode:       "blink",
		PromptStyle:      gumstyle.Styles{},
		PlaceholderStyle: gumstyle.Styles{Foreground: "240"},
		CursorStyle:      gumstyle.Styles{Foreground: "212"},
		HeaderStyle:      gumstyle.Styles{Foreground: "240"},
	}
	out, err := captureStdout(opts.Run)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func captureStdout(fn func() error) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stdout
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	if runErr != nil {
		return "", runErr
	}
	return buf.String(), nil
}
