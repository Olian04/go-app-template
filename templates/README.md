# [[.ModuleBasename]]

Generated Go module (`[[.ModulePath]]`, Go [[.GoVersion]]). Mode: `[[.Mode]]`.

## Layout

| Path | Role |
| --- | --- |
| `internal/domain/echo` | Demo domain model — the same in every mode. |
[[- if modeIs "cli" "cli-library" ]]
| `cmd/[[.CliName]]` | CLI adapter: args/stdin → domain → stdout. |
[[- end ]]
[[- if modeIs "http" ]]
| `cmd/[[.ServiceName]]` | Service entrypoint. |
[[- end ]]
[[- if modeIs "library" "cli-library" ]]
| `pkg/[[.LibName]]` | Public library facade over the domain. |
[[- end ]]
[[- if modeIs "http" ]]
| `internal/app` | Composition root; owns listener timeouts + graceful shutdown. |
| `internal/transport/http` | HTTP adapter: request → domain → JSON, plus middleware. |
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
| `test/unit/` | Unit tests mirroring [[ if modeIs "library" ]]`internal/` / `pkg/`[[ else if modeIs "cli-library" ]]`internal/` / `pkg/`[[ else ]]`internal/`[[ end ]]. |

## Dependency direction

The domain is the fixed point; everything else adapts to it and depends inward.
[[ if modeIs "library" ]]
`pkg/[[.LibName]]` → `internal/domain/echo`. The facade exists because consumers
outside this module cannot import `internal/`; it translates plain Go types to
domain types. Keep it importable without side effects: no global state, no `init()`.
[[ else if modeIs "http" ]]
`cmd` → `internal/app` → domain, transports, observability, config.
Domain avoids HTTP / slog / Prometheus imports.
[[ else ]]
`cmd` → domain[[ if modeIs "cli-library" ]], `pkg/[[.LibName]]`[[ end ]], observability, config.
There is no `internal/app` in this mode; `cmd` is the composition root, and should
only parse input and wire dependencies — behaviour belongs in the domain.
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
[[- if modeIs "cli" "cli-library" ]]

Reach the domain through the CLI. Logs go to stderr, so stdout stays pipeable:

```bash
make build
./dist/[[.CliName]] hello world          # args
echo '  padded  ' | ./dist/[[.CliName]]  # stdin
./dist/[[.CliName]] hello 2>/dev/null    # result only
```
[[- end ]]
[[- if modeIs "library" ]]

Reach the domain through the exported facade:

```go
import "[[.ModulePath]]/pkg/[[.LibName]]"

func main() {
	println([[.LibName]].Echo("  hello  ")) // "hello"
}
```
[[- end ]]
[[- if modeIs "cli-library" ]]

The library facade reaches the same domain the CLI does:

```go
import "[[.ModulePath]]/pkg/[[.LibName]]"

func main() {
	println([[.LibName]].Echo("  hello  ")) // "hello"
}
```
[[- end ]]
[[- if modeIs "http" ]]

Reach the domain over HTTP:

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
