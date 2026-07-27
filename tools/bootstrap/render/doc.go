// Package render implements dual-mode [[ ]] template rendering.
//
// Gate line 1: [[ when (modeIs "http") ]], [[ when (modeIs "cli" "cli-library") ]].
// Parentheses required (standard nesting). No rewrite of bare when modeIs.
// Gate FuncMap: when, modeIs. Gate data = {Mode}.
// Body FuncMap: modeIs only (no when / whenEither / whenMode); full Context.
//
// API (frozen by WAVE0):
//
//	type Mode string
//
//	const (
//	    ModeCLI        Mode = "cli"
//	    ModeLibrary    Mode = "library"
//	    ModeCLILibrary Mode = "cli-library"
//	    ModeHTTP       Mode = "http"
//	)
//
//	type Binary struct {
//	    Name, MainPackage, VersionPackage string
//	}
//
//	type Context struct {
//	    Mode Mode
//	    ModulePath, ModuleBasename, GoVersion string
//	    MetricPrefix string // Prometheus-safe form of ModuleBasename
//	    LibName, CliName, ServiceName string // LibName = Go package id [a-z][a-z0-9]*
//	    Binary *Binary // nil when mode has no binary (library)
//	}
//
//	// File runs gate+normal pipeline for one file; skip => skipped=true.
//	func File(path string, src []byte, ctx Context) (out []byte, skipped bool, err error)
//
//	// Tree walks templatesDir → outDir; expands path templates; duplicate out path → error.
//	func Tree(templatesDir, outDir string, ctx Context) error
package render

// Mode is the bootstrap product shape.
type Mode string

const (
	ModeCLI        Mode = "cli"
	ModeLibrary    Mode = "library"
	ModeCLILibrary Mode = "cli-library"
	ModeHTTP       Mode = "http"
)

// Binary describes one built binary entry for release/ldflags templates.
type Binary struct {
	Name           string
	MainPackage    string
	VersionPackage string
}

// Context is the normal-mode template data root.
type Context struct {
	Mode           Mode
	ModulePath     string
	ModuleBasename string
	GoVersion      string
	MetricPrefix   string // Prometheus-safe form of ModuleBasename (hyphens → _)
	LibName        string // Go package identifier ([a-z][a-z0-9]*); no hyphens
	CliName        string
	ServiceName    string
	Binary         *Binary // nil when mode has no binary (library)
}
