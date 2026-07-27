# [[.ModuleBasename]]

Generated Go module (`[[.ModulePath]]`, Go [[.GoVersion]]). Mode: `[[.Mode]]`.

## Layout

| Path | Role |
| --- | --- |
[[- if modeIs "cli" "cli-library" "http" ]]
| `cmd/<binary>` | Entrypoints. |
[[- end ]]
[[- if modeIs "library" "cli-library" ]]
| `pkg/[[.LibName]]` | Public library surface. |
[[- end ]]
[[- if modeIs "http" ]]
| `internal/app` | Composition root + HTTP wiring. |
| `internal/domain` | Pure domain behavior. |
| `internal/transport/http` | HTTP routes. |
[[- end ]]
[[- if modeIs "cli" "cli-library" "http" ]]
| `internal/config` | Config load (YAML → ENV → flags). |
| `configs/` | Example config files. |
| `internal/observability/logging` | slog setup helpers. |
[[- end ]]
[[- if modeIs "http" ]]
| `internal/observability/metrics` | Prometheus registry. |
| `internal/transport/metricshttp` | `/metrics` server helper. |
[[- end ]]
| `test/unit/` | Unit tests mirroring `internal/` / `pkg/`. |

## Dependency direction

`cmd` → `internal/app` → domain, transports[[- if modeIs "cli" "cli-library" "http" ]], observability[[- end ]][[- if modeIs "cli" "cli-library" "http" ]], config[[- end ]]. Domain avoids HTTP / slog / Prometheus imports.

[[- if modeIs "cli" "cli-library" "http" ]]
## Config

Precedence: YAML → ENV → flags.
Point `APP_CONFIG_FILE` at a YAML file (omit for built-in defaults). See `configs/`.
[[- end ]]

## Run locally

```bash
make help
[[- if modeIs "cli" "cli-library" "http" ]]
make run
[[- end ]]
```

[[- if modeIs "http" ]]
```bash
curl -X POST http://localhost:8080/echo -H 'Content-Type: application/json' -d '{"message":" hello "}'
curl http://localhost:9090/metrics
```
[[- end ]]

## Checks

```bash
make format
make lint
make test
make test-race
```

## Releases

[[- if modeIs "cli" "cli-library" "http" ]]
Push tag `v*`; CI runs GoReleaser for binaries.
[[- end ]]
[[- if modeIs "http" ]]
Docker image build via goreleaser/Dockerfile.
[[- end ]]
[[- if modeIs "library" "cli-library" ]]
Library release artifacts included for library modes.
[[- end ]]

## Agent orientation

See `docs/` and `CLAUDE.md` for agent context.
