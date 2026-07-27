package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Olian04/go-app-template/tools/bootstrap/render"
)

type selection struct {
	Mode        render.Mode
	ModulePath  string
	GoVersion   string
	CliName     string
	ServiceName string
	LibName     string
}

var (
	reModulePath  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._\-/]*[A-Za-z0-9]$`)
	reName        = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`) // binary / dir names
	rePackageName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)   // Go package identifier
)

func parseMode(s string) (render.Mode, error) {
	m := render.Mode(strings.TrimSpace(s))
	switch m {
	case render.ModeCLI, render.ModeLibrary, render.ModeCLILibrary, render.ModeHTTP:
		return m, nil
	case "":
		return "", fmt.Errorf("mode required (cli|library|cli-library|http)")
	default:
		return "", fmt.Errorf("invalid mode %q (want cli|library|cli-library|http)", s)
	}
}

func modeNeedsCLI(m render.Mode) bool {
	return m == render.ModeCLI || m == render.ModeCLILibrary
}

func modeNeedsLibrary(m render.Mode) bool {
	return m == render.ModeLibrary || m == render.ModeCLILibrary
}

func modeNeedsHTTP(m render.Mode) bool {
	return m == render.ModeHTTP
}

func applyNameDefaults(sel *selection) {
	base := moduleBasename(sel.ModulePath)
	if modeNeedsCLI(sel.Mode) && sel.CliName == "" {
		sel.CliName = base
	}
	if modeNeedsHTTP(sel.Mode) && sel.ServiceName == "" {
		sel.ServiceName = base
	}
	if modeNeedsLibrary(sel.Mode) && sel.LibName == "" {
		sel.LibName = goPackageName(base)
	}
}

func buildContext(sel selection) (render.Context, error) {
	base := moduleBasename(sel.ModulePath)
	ctx := render.Context{
		Mode:           sel.Mode,
		ModulePath:     strings.TrimSpace(sel.ModulePath),
		ModuleBasename: base,
		GoVersion:      strings.TrimSpace(sel.GoVersion),
		MetricPrefix:   prometheusMetricPrefix(base),
		LibName:        strings.TrimSpace(sel.LibName),
		CliName:        strings.TrimSpace(sel.CliName),
		ServiceName:    strings.TrimSpace(sel.ServiceName),
	}

	switch sel.Mode {
	case render.ModeHTTP:
		b := binaryEntry(ctx.ModulePath, ctx.ServiceName)
		ctx.Binary = &b
	case render.ModeCLI, render.ModeCLILibrary:
		b := binaryEntry(ctx.ModulePath, ctx.CliName)
		ctx.Binary = &b
	}
	return ctx, nil
}

func binaryEntry(modulePath, name string) render.Binary {
	mainPkg := modulePath + "/cmd/" + name
	return render.Binary{
		Name:           name,
		MainPackage:    mainPkg,
		VersionPackage: mainPkg + "/version",
	}
}

func validateContext(ctx render.Context) error {
	if ctx.Mode == "" {
		return fmt.Errorf("mode required (cli|library|cli-library|http)")
	}
	if _, err := parseMode(string(ctx.Mode)); err != nil {
		return err
	}
	if ctx.ModulePath == "" {
		return fmt.Errorf("module path required (set --module-path or MODULE_PATH)")
	}
	if !reModulePath.MatchString(ctx.ModulePath) || strings.Contains(ctx.ModulePath, "//") {
		return fmt.Errorf("invalid module path %q", ctx.ModulePath)
	}
	if ctx.GoVersion == "" {
		return fmt.Errorf("go version required")
	}
	if modeNeedsCLI(ctx.Mode) {
		if err := validateName("cli-name", ctx.CliName); err != nil {
			return err
		}
	}
	if modeNeedsHTTP(ctx.Mode) {
		if err := validateName("service-name", ctx.ServiceName); err != nil {
			return err
		}
	}
	if modeNeedsLibrary(ctx.Mode) {
		if err := validatePackageName("lib-name", ctx.LibName); err != nil {
			return err
		}
	}
	return nil
}

func validateName(field, name string) error {
	if name == "" {
		return fmt.Errorf("%s required", field)
	}
	if !reName.MatchString(name) {
		return fmt.Errorf("invalid %s %q (want lowercase [a-z][a-z0-9_-]*)", field, name)
	}
	return nil
}

func validatePackageName(field, name string) error {
	if name == "" {
		return fmt.Errorf("%s required", field)
	}
	if !rePackageName.MatchString(name) {
		return fmt.Errorf("invalid %s %q (want Go package name [a-z][a-z0-9]*; no hyphens)", field, name)
	}
	return nil
}

func moduleBasename(modulePath string) string {
	base := filepath.Base(strings.TrimSuffix(modulePath, "/"))
	base = strings.TrimSuffix(base, ".git")
	return strings.ToLower(base)
}

// goPackageName maps a basename to a Go package identifier: lowercase [a-z][a-z0-9]*.
// Hyphens/underscores/other punctuation are stripped (Go style: single-word names).
func goPackageName(basename string) string {
	basename = strings.ToLower(strings.TrimSpace(basename))
	var b strings.Builder
	b.Grow(len(basename))
	for _, r := range basename {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "lib"
	}
	if out[0] < 'a' || out[0] > 'z' {
		out = "lib" + out
	}
	return out
}

// prometheusMetricPrefix maps a module basename to a legacy Prometheus metric
// name prefix ([a-zA-Z_:][a-zA-Z0-9_:]*). Hyphens and other invalid runes become '_'.
func prometheusMetricPrefix(basename string) string {
	basename = strings.TrimSpace(basename)
	if basename == "" {
		return "app"
	}
	var b strings.Builder
	b.Grow(len(basename))
	for _, r := range basename {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == ':':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "app"
	}
	first := out[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_' || first == ':') {
		out = "app_" + out
	}
	return out
}

func defaultGoVersion() string {
	if v := strings.TrimSpace(os.Getenv("GO_VERSION")); v != "" {
		v = strings.TrimPrefix(v, "go")
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
		return v
	}
	v := strings.TrimPrefix(runtime.Version(), "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		templates := filepath.Join(dir, "templates")
		bootstrap := filepath.Join(dir, "tools", "bootstrap")
		if dirExists(templates) && dirExists(bootstrap) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("repo root not found from %s (need templates/ + tools/bootstrap/)", wd)
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func defaultModulePath() (string, error) {
	if v := strings.TrimSpace(os.Getenv("MODULE_PATH")); v != "" {
		return v, nil
	}
	return remoteModulePath()
}

func remoteModulePath() (string, error) {
	root, err := findRepoRoot()
	if err != nil {
		return "", err
	}
	raw, err := gitRemoteOrigin(root)
	if err != nil {
		return "", err
	}
	return normalizeRemoteURL(raw), nil
}

func gitRemoteOrigin(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, ".git", "config"))
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(b), "\n")
	inOrigin := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "[") {
			inOrigin = trim == `[remote "origin"]`
			continue
		}
		if inOrigin && strings.HasPrefix(trim, "url") {
			parts := strings.SplitN(trim, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return "", fmt.Errorf("git remote origin not found")
}

func normalizeRemoteURL(remote string) string {
	remote = strings.TrimSuffix(strings.TrimSpace(remote), ".git")
	switch {
	case strings.HasPrefix(remote, "git@"):
		remote = strings.TrimPrefix(remote, "git@")
		remote = strings.Replace(remote, ":", "/", 1)
	case strings.HasPrefix(remote, "ssh://git@"):
		remote = strings.TrimPrefix(remote, "ssh://git@")
	case strings.HasPrefix(remote, "https://"):
		remote = strings.TrimPrefix(remote, "https://")
	case strings.HasPrefix(remote, "http://"):
		remote = strings.TrimPrefix(remote, "http://")
	}
	return remote
}
