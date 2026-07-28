# Bootstrap shared contracts (WAVE0 freeze)

Source of truth for WAVE1 agents. Do not rename fields or change gate dialect without a new WAVE0.

## Delimiters

- Open: `[[`
- Close: `]]`

## Mode values

`cli` | `library` | `cli-library` | `http`

## Types (package `render`)

```go
type Mode string

const (
    ModeCLI        Mode = "cli"
    ModeLibrary    Mode = "library"
    ModeCLILibrary Mode = "cli-library"
    ModeHTTP       Mode = "http"
)

type Binary struct{ Name, MainPackage, VersionPackage string }

type Context struct {
    Mode Mode
    ModulePath, ModuleBasename, GoVersion string
    MetricPrefix string // Prometheus-safe form of ModuleBasename (hyphens → _)
    LibName, CliName, ServiceName string
    Binary *Binary // nil when mode has no binary (library)
}

// Naming:
// - CliName / ServiceName: binary + cmd/ dir ([a-z][a-z0-9_-]*); hyphens OK
// - LibName: Go package identifier ([a-z][a-z0-9]*); no hyphens/underscores
//   (default from module basename strips punctuation: go-app-template → goapptemplate)

```

Mode alone drives gates. No feature flags / `.Include.*`.

## Gate dialect

- Line 1 may wrap gate in comments.
- Primary form: `[[ when (modeIs "http") ]]` or `[[ when (modeIs "cli" "cli-library") ]]`
- Parentheses required (standard Go `text/template` nesting). Bare `when modeIs …` is invalid — no rewriter invents them.
- Gate FuncMap: `when`, `modeIs`; data `{ Mode Mode }`
- `when` takes exactly one bool (from `modeIs`)
- `modeIs` args = allowed mode strings; true if current mode is any of them (OR)
- Body FuncMap: **`modeIs` only** — no `when` / `whenEither` / `whenMode` (using them in body → error); full `Context`
- Gate execute data = `{ Mode Mode }` only (dual-mode isolation; no `ModulePath` / prompts)
- No `whenEither` / `whenMode` (removed; mode OR is `modeIs` multi-arg)

### Gate mode (line 1 only)

- Gate may be **embedded** in line 1 (language comments OK); whole line need not be only `[[ when … ]]`
- Detection: line 1 contains a `[[ … ]]` action whose leading ident is `when`
- Extract that `[[ … ]]` substring and execute it alone (comment wrappers ignored during eval)
- Multiple gate actions on one line → hard error
- Hyphen names forbidden (`when-mode` invalid) → hard error
- Gate-looking `[[…]]` that fails gate mode → hard error (do not silently include)
- No gate `when` on line 1 → ungated (line 1 kept; normal mode on full file)
- Plain line-1 `[[ … ]]` that is only a template action and not `when` → hard error

Examples:

```text
# [[ when (modeIs "http") ]]
// [[ when (modeIs "cli" "cli-library") ]]
/* [[ when (modeIs "library" "cli-library") ]] */
[[ when (modeIs "cli" "cli-library" "http") ]]
<!-- [[ when (modeIs "cli" "library" "cli-library" "http") ]] -->
```

On true: strip **entire** first line (including comment wrapper), then normal-mode render remainder.
On false: skip file. No merge. Prune empty dirs after Tree.

### Mode cheat sheet for templates

| Files                                    | Gate / body                                               |
| ---------------------------------------- | --------------------------------------------------------- |
| Domain model + its tests                 | ungated (all modes) — the shared demo every adapter wraps  |
| CLI cmd                                  | `when (modeIs "cli" "cli-library")`                         |
| Library pkg (facade over domain)         | `when (modeIs "library" "cli-library")`                     |
| HTTP cmd/app/transport/middleware        | `when (modeIs "http")`                                      |
| Config root + logging + labels + yaml    | `when (modeIs "cli" "cli-library" "http")`                  |
| Logging correlation-id helpers           | `when (modeIs "cli" "cli-library" "http")`                  |
| Config `http` section (timeouts, limits) | `when (modeIs "http")`                                      |
| Metrics config + registry + metricshttp  | `when (modeIs "http")`                                      |
| Release docker bits                      | body/`[[ if modeIs "http" ]]`                             |
| Release library bits                     | `when (modeIs "library" "cli-library")`                     |
| Release binary bits                      | `when (modeIs "cli" "cli-library" "http")`                  |
| CI / AiRules                             | ungated (always) **or** all four modes—**prefer ungated** |

The user-facing feature×mode matrix in the repo `README.md` is the same mapping
in prose; keep both in step.

### Gate/body hygiene

- A body conditional whose mode set is a **superset** of the file's gate is always
  true. Drop it — a reader must be able to trust that a visible `[[ if ]]` can
  actually be false. Example: inside a file gated
  `when (modeIs "cli" "cli-library" "http")`, a body `[[ if modeIs "cli" "cli-library" "http" ]]`
  is noise, and an `[[ else ]]` on it is dead code.
- A body conditional **disjoint** from the gate is dead code; delete the block.
- Prefer narrowing the file gate over wrapping the whole body in one conditional.
- `[[ if ]]` on its own line leaves a blank line and can misalign struct tags.
  That is fine for Go: bootstrap gofmts the rendered tree (`formatGoFiles`), so
  templates stay readable. Non-Go files need explicit `[[- ... -]]` trim markers.

### Enforcement

`tools/bootstrap/manifest_test.go` renders all four modes and diffs the file set
against `testdata/manifest/<mode>.txt`, asserts named per-file mode rules, and
rejects leftover `[[`/`]]` in output. After an intentional gate change:

```bash
go -C tools/bootstrap test . -run TestModeManifest -update
```

Review the manifest diff — it is the mode-enablement change in reviewable form.

## Render API

```go
package render

// File: gate+normal pipeline for one file; skip => skipped=true
func File(path string, src []byte, ctx Context) (out []byte, skipped bool, err error)

// Tree: walk templatesDir → outDir; path templates; duplicate out path → error
func Tree(templatesDir, outDir string, ctx Context) error
```

UX imports `render` and calls `Tree`. WAVE0 may stub File/Tree; WAVE1-render owns `when`/`modeIs` + real gate pipeline.

## Module

```text
github.com/Olian04/go-app-template/tools/bootstrap
```

Go 1.26. Nested module under `tools/bootstrap/`.
