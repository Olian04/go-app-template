# Go app template

In-place Go project template. After you use this repository as a GitHub template (or clone it), run bootstrap to choose a mode and rewrite the tree into your app.

## Modes

| Mode | What you get |
| --- | --- |
| `cli` | CLI binary |
| `library` | Reusable Go package |
| `cli-library` | CLI + library |
| `http` | HTTP service |

Mode is the **only** switch: it decides which files render and which features exist.
There are no runtime feature flags.

Every mode ships the same demo domain model (`internal/domain/echo`) and differs only
in the IO wrapped around it — a CLI command, an HTTP handler, or an exported library
facade. That shows where your own logic goes and what changes when the mode does.

### What each mode enables

| Feature | `cli` | `library` | `cli-library` | `http` |
| --- | :-: | :-: | :-: | :-: |
| Demo domain model (`internal/domain/echo/`) | ✓ | ✓ | ✓ | ✓ |
| CLI binary (`cmd/<cli-name>/`) — args/stdin → domain → stdout | ✓ | – | ✓ | – |
| HTTP service binary (`cmd/<service-name>/`) — request → domain → JSON | – | – | – | ✓ |
| Public library package (`pkg/<lib-name>/`) — exported facade over domain | – | ✓ | ✓ | – |
| Layered config: `labels`, `logging` (`internal/config/`) | ✓ | – | ✓ | ✓ |
| Config `http` section — timeouts, body/header limits | – | – | – | ✓ |
| Config `metrics` section | – | – | – | ✓ |
| slog setup (`internal/observability/logging/`) | ✓ | – | ✓ | ✓ |
| Correlation-ID context helpers (`logging.FromContext`) | ✓ | – | ✓ | ✓ |
| HTTP transport + composition root (`internal/transport/`, `internal/app/`) | – | – | – | ✓ |
| Middleware chain: recover, request ID, logging, metrics | – | – | – | ✓ |
| `X-Request-ID` propagation on requests/responses | – | – | – | ✓ |
| Prometheus registry + `/metrics` listener | – | – | – | ✓ |
| Request rate/error/duration + in-flight metrics | – | – | – | ✓ |
| Server timeouts, `MaxHeaderBytes`, `MaxBytesReader` | – | – | – | ✓ |
| Coordinated graceful shutdown of all listeners | – | – | – | ✓ |
| `make build` / `make run` | ✓ | – | ✓ | ✓ |
| CI workflow, lint config, agent rules, license | ✓ | ✓ | ✓ | ✓ |
| GoReleaser binaries + SBOM | ✓ | – | ✓ | ✓ |
| GoReleaser multi-arch Docker image | – | – | – | ✓ |

Legend: ✓ enabled, – not rendered.

This table is enforced, not aspirational: `tools/bootstrap/testdata/manifest/<mode>.txt`
records the exact file set per mode, and `TestModeManifest` fails when a file
renders in the wrong mode. Re-record intentional changes with:

```bash
go -C tools/bootstrap test . -run TestModeManifest -update
```

## Bootstrap

Requires Go 1.26+ on `PATH`.

Interactive (gum prompts). Mode choose, then module path and names — prefilled from git remote / basename when available:

```bash
./bootstrap.sh
```

Noninteractive — pass `-mode=` (and optionally names; otherwise git-inferred defaults apply):

```bash
./bootstrap.sh \
  -noninteractive \
  -mode=http \
  -module-path=example.com/you/your-app \
  -service-name=yourapp
```

Bootstrap renders `templates/` into a staging dir, swaps that tree into the repo root, runs `go mod tidy`, then removes `templates/`, `tools/bootstrap/`, and `bootstrap.sh` unless you keep them:

```bash
BOOTSTRAP_KEEP=1 ./bootstrap.sh ...
# or
./bootstrap.sh -keep ...
```

Useful flags:

| Flag / env | Meaning |
| --- | --- |
| `-mode` | `cli` \| `library` \| `cli-library` \| `http` |
| `-noninteractive` / `BOOTSTRAP_NONINTERACTIVE=1` | Skip prompts; use flags/env |
| `-dry-run` | Print render context JSON; no files written |
| `-no-swap` | Render to `.bootstrap-out` only |
| `-keep` / `BOOTSTRAP_KEEP=1` | Keep bootstrap tooling after swap |
| `-module-path` / `MODULE_PATH` | Go module path (else git remote) |
| `-cli-name` / `-lib-name` / `-service-name` | Binary / package / service names (else git basename) |

## Smoke (maintainer)

From a **temp copy** of the working tree (do not wipe the template checkout you develop in):

```bash
SMOKE_DIR=$(mktemp -d)
rsync -a --exclude '.git' --exclude '.bootstrap-out' ./ "$SMOKE_DIR/"
(
  cd "$SMOKE_DIR"
  ./bootstrap.sh \
    -noninteractive \
    -mode=http \
    -module-path=example.com/smoke/app \
    -service-name=smokeapp
  make test
  make build
)
rm -rf "$SMOKE_DIR"
```

Exit code of that block should be `0`. (After changes are committed, `git worktree add` is also fine.)

Prefer `make test` / `make build` over bare `go` commands here: `SOURCE_CODE` in the
generated `Makefile` is itself mode-dependent, so only the Make targets catch a
target that references a tree the mode does not render.

Repeat for `cli`, `library`, and `cli-library` — or let `.github/workflows/template-ci.yml`
do it, which smoke-runs all four modes and asserts each mode's surface.

To inspect a mode without touching the checkout, render to staging only:

```bash
./bootstrap.sh -noninteractive -no-swap -mode=cli -module-path=example.com/x/app
ls .bootstrap-out
```

## Layout (this template repo)

| Path | Role |
| --- | --- |
| `bootstrap.sh` | Thin launcher → `go -C tools/bootstrap run .` |
| `tools/bootstrap/` | UX + dual-mode `[[ ]]` render engine |
| `tools/bootstrap/contract.md` | Gate dialect + mode cheat sheet (authoring reference) |
| `tools/bootstrap/testdata/manifest/` | Golden per-mode file lists |
| `templates/` | Mode-gated product templates |
| `README.md` | This file |

After bootstrap on a generated project, the product tree (`cmd/`, `internal/`, `go.mod`, …) replaces the template machinery.
