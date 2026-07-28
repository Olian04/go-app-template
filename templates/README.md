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
| `internal/app` | Composition root; owns listener timeouts + graceful shutdown. |
| `internal/domain` | Pure domain behavior. |
| `internal/transport/http` | HTTP routes + middleware chain. |
[[- end ]]
[[- if modeIs "cli" "cli-library" "http" ]]
| `internal/config` | Config load (YAML → ENV → flags). |
| `configs/` | Example config files. |
| `internal/observability/logging` | slog setup + request-ID context helpers. |
[[- end ]]
[[- if modeIs "http" ]]
| `internal/observability/metrics` | Prometheus registry (request, duration, in-flight). |
| `internal/transport/metricshttp` | `/metrics` handler. |
[[- end ]]
| `test/unit/` | Unit tests mirroring [[ if modeIs "library" ]]`pkg/`[[ else if modeIs "cli-library" ]]`internal/` / `pkg/`[[ else ]]`internal/`[[ end ]]. |

## Dependency direction
[[ if modeIs "library" ]]
`pkg/[[.LibName]]` is the whole public surface and depends only on the standard library.
Keep it importable without side effects: no global state, no `init()` work.
[[ else if modeIs "http" ]]
`cmd` → `internal/app` → domain, transports, observability, config.
Domain avoids HTTP / slog / Prometheus imports.
[[ else ]]
`cmd` → [[ if modeIs "cli-library" ]]`pkg/[[.LibName]]`, [[ end ]]observability, config.
Keep business logic out of `cmd`; it should only parse flags and wire dependencies.
[[ end ]]

[[- if modeIs "cli" "cli-library" "http" ]]
## Config

Precedence: YAML → ENV → flags.
Point `APP_CONFIG_FILE` at a YAML file (omit for built-in defaults). See `configs/`.

Sections in this mode: `labels`[[ if modeIs "http" ]], `http`, `metrics`[[ end ]], `logging`.
[[- end ]]
[[- if modeIs "http" ]]

## Observability & hardening

- Every response carries `X-Request-ID`; an inbound one is reused so traces span services.
  Log with `logging.FromContext(ctx)` to inherit `request_id`.
- `/metrics` exposes request totals, latency histogram, and an in-flight gauge.
- Server timeouts, header/body caps, and the shutdown grace period are config
  (`http:` section). Requests over `max_body_bytes` get `413`; unknown JSON fields get `400`.
- `SIGINT`/`SIGTERM` drain both listeners before exit.
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
# -i to see the X-Request-ID response header
curl -i -X POST http://localhost:8080/echo -H 'Content-Type: application/json' -d '{"message":" hello "}'
curl -X POST http://localhost:8080/echo -H 'X-Request-ID: my-trace' -d '{"message":"hi"}'
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
