package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		nonInteractive = fs.Bool("noninteractive", envBool("BOOTSTRAP_NONINTERACTIVE"), "skip prompts; use flags/env")
		dryRun         = fs.Bool("dry-run", false, "print render context as JSON and exit")
		modeFlag       = fs.String("mode", envOr("BOOTSTRAP_MODE", ""), "product mode: cli|library|cli-library|http")
		modulePath     = fs.String("module-path", envOr("MODULE_PATH", ""), "Go module path")
		goVersion      = fs.String("go-version", envOr("GO_VERSION", ""), "Go major.minor (default: runtime)")
		cliName        = fs.String("cli-name", firstNonEmpty(envOr("CLI_NAME", ""), envOr("CMD_NAME", "")), "CLI binary/cmd name")
		serviceName    = fs.String("service-name", envOr("SERVICE_NAME", ""), "HTTP service cmd name")
		libName        = fs.String("lib-name", envOr("LIB_NAME", ""), "library Go package name ([a-z][a-z0-9]*; no hyphens)")
		outDir         = fs.String("out", envOr("BOOTSTRAP_OUT", ""), "staging output directory")
		templatesDir   = fs.String("templates", envOr("BOOTSTRAP_TEMPLATES", ""), "templates directory")
		noSwap         = fs.Bool("no-swap", false, "render to staging only; skip swap/tidy/wipe")
		keep           = fs.Bool("keep", envBool("BOOTSTRAP_KEEP"), "keep templates/, tools/bootstrap/, bootstrap.sh after swap")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	if *templatesDir == "" {
		*templatesDir = filepath.Join(root, "templates")
	}
	if *outDir == "" {
		*outDir = filepath.Join(root, ".bootstrap-out")
	}

	base := selection{
		ModulePath:  *modulePath,
		GoVersion:   *goVersion,
		CliName:     *cliName,
		ServiceName: *serviceName,
		LibName:     *libName,
	}
	if *modeFlag != "" {
		mode, err := parseMode(*modeFlag)
		if err != nil {
			return err
		}
		base.Mode = mode
	}

	var sel selection
	if *nonInteractive || *modeFlag != "" || !isInteractive() {
		if base.Mode == "" {
			return fmt.Errorf("mode required (pass -mode=cli|library|cli-library|http)")
		}
		sel = base
	} else {
		sel, err = promptInteractive(base)
		if err != nil {
			return err
		}
	}

	if sel.GoVersion == "" {
		sel.GoVersion = defaultGoVersion()
	}
	if sel.ModulePath == "" {
		if hint, hintErr := defaultModulePath(); hintErr == nil {
			sel.ModulePath = hint
		}
	}
	applyNameDefaults(&sel)

	ctx, err := buildContext(sel)
	if err != nil {
		return err
	}
	if err := validateContext(ctx); err != nil {
		return err
	}

	if *dryRun {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(ctx)
	}

	if err := os.RemoveAll(*outDir); err != nil {
		return fmt.Errorf("clear staging: %w", err)
	}
	if err := render.Tree(*templatesDir, *outDir, ctx); err != nil {
		return fmt.Errorf("render tree: %w", err)
	}
	if err := formatGoFiles(*outDir); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "bootstrap: rendered into %s\n", *outDir)

	if *noSwap {
		return nil
	}

	fmt.Fprintln(os.Stderr, "bootstrap: swapping staging into repo root")
	if err := swapStaging(root, *outDir); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "bootstrap: running go mod tidy")
	if err := tidyModule(root); err != nil {
		return err
	}

	if *keep {
		fmt.Fprintln(os.Stderr, "bootstrap: keeping templates/ tools/bootstrap/ bootstrap.sh (BOOTSTRAP_KEEP/--keep)")
		return nil
	}

	fmt.Fprintln(os.Stderr, "bootstrap: wiping templates/ tools/bootstrap/ bootstrap.sh")
	return wipeBootstrap(root)
}
