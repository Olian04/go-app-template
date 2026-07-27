# Go app template

In-place Go project template. After you use this repository as a GitHub template (or clone it), run bootstrap to choose a mode and rewrite the tree into your app.

## Modes

| Mode | What you get |
| --- | --- |
| `cli` | CLI binary |
| `library` | Reusable Go package |
| `cli-library` | CLI + library |
| `http` | HTTP service |

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
  go test ./...
  go build ./...
)
rm -rf "$SMOKE_DIR"
```

Exit code of that block should be `0`. (After changes are committed, `git worktree add` is also fine.)

## Layout (this template repo)

| Path | Role |
| --- | --- |
| `bootstrap.sh` | Thin launcher → `go -C tools/bootstrap run .` |
| `tools/bootstrap/` | UX + dual-mode `[[ ]]` render engine |
| `templates/` | Mode-gated product templates |
| `README.md` | This file |

After bootstrap on a generated project, the product tree (`cmd/`, `internal/`, `go.mod`, …) replaces the template machinery.
