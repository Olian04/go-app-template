# Agent context: [[ .ModuleBasename ]]

App template — use with repo `README.md`[[ if modeIs "cli" "cli-library" "http" ]] and [`configs/config.example.yaml`](../configs/config.example.yaml)[[ end ]]. Mode: `[[ .Mode ]]`.

## Layout

| Path | Role |
| --- | --- |
| `internal/domain/echo` | Domain logic only. Identical in every mode; no IO imports. |
[[ if modeIs "http" ]]
| `cmd/[[ .ServiceName ]]` | Entry: load config, `logging.Setup`, signal context, `app.Run`. |
| `internal/app` | Wire registry + router; own every listener's timeouts and shutdown. |
| `internal/transport/http` | HTTP adapter over the domain + middleware chain (recover, request ID, logging, metrics). |
| `internal/transport/metricshttp` | `/metrics` handler (no lifecycle — `internal/app` runs it). |
[[ end ]]
[[ if modeIs "cli" "cli-library" ]]
| `cmd/[[ .CliName ]]` | CLI adapter over the domain: args/stdin in, stdout out. |
[[ end ]]
[[ if modeIs "library" "cli-library" ]]
| `pkg/[[ .LibName ]]` | Public API: exported facade delegating to the domain. |
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
| `internal/config` | Root config `Load`; section structs (defaults/`WithDefaults`/`Validate`). |
| `internal/observability/logging` | slog setup + request-ID context helpers. |
| `configs` | YAML contract examples. |
[[ end ]]
[[ if modeIs "http" ]]
| `internal/observability/metrics` | Prometheus registry + handler. |
[[ end ]]
| `test/unit/...` | Unit tests beside mirrored paths. |

## Dependency direction

`internal/domain` is the fixed point. Mode-specific IO adapters depend on it; it
depends on nothing but the standard library. Put behaviour in the domain and
keep adapters to translation only.
[[ if modeIs "library" ]]
`pkg/[[ .LibName ]]` → `internal/domain/echo`. The facade is the public surface
(consumers cannot import `internal/`); no side effects on import.
[[ else if modeIs "http" ]]
`cmd` → `internal/app` → domain, transport, observability, aggregated `internal/config`.
Domain must not import app, transports, observability.
[[ else ]]
`cmd` → `internal/domain/echo`[[ if modeIs "cli-library" ]], `pkg/[[ .LibName ]]`[[ end ]], observability, aggregated `internal/config`.
There is no `internal/app` in this mode; `cmd` is the composition root.
[[ end ]]

[[ if modeIs "cli" "cli-library" "http" ]]
## Configuration

`APP_CONFIG_FILE` optional; unset uses defaults matching example file shapes.

Section structs carry `Default*` constants in-package, `WithDefaults()`, `Validate()`. Root `config.Config.Validate()` delegates down.

Sections in this mode: `labels`[[ if modeIs "http" ]], `http`, `metrics`[[ end ]], `logging`. Mode decides
which sections exist — do not add an `http` section to a mode with no server.
[[ end ]]

## Mode notes

[[ if modeIs "cli" "cli-library" "http" ]]
- **Logging**: use `internal/observability/logging` for slog setup. Prefer
  `logging.FromContext(ctx)` over package-level `slog` so lines carry `request_id`.
- **Exit errors**: report from `main` to stderr, not `slog` — the configured
  logger is already torn down when the deferred cleanup has run.
[[ else ]]
- **Logging**: standard library `slog` defaults (no `observability/logging` package).
[[ end ]]
[[ if modeIs "http" ]]
- **Metrics**: Prometheus registry + `/metrics` handler; `internal/app` serves it.
  Middleware records request count, duration, and in-flight gauge.
- **Request IDs**: `RequestID` middleware honours inbound `X-Request-ID` (sanitized)
  or generates one, echoes it, and puts it on the context.
- **Hardening**: server timeouts and `MaxHeaderBytes` come from config; handlers
  bound bodies with `http.MaxBytesReader` and reject unknown JSON fields. Do not
  return decoder errors to clients — log them and send a fixed message.
- **Shutdown**: `app.Run` cancels siblings on first failure and waits for every
  listener to drain before returning.
[[ else ]]
- **Metrics**: off for this mode.
[[ end ]]

## Commands (`make`)

`lint`, `test`, `test-race`[[ if modeIs "http" ]], `run`, `build` (output `./dist/[[ .ServiceName ]]`)[[ end ]][[ if modeIs "cli" "cli-library" ]], `run`, `build` (output `./dist/[[ .CliName ]]`)[[ end ]].

---

## Go proverbs ([source](https://go-proverbs.github.io/), caveman compress)

Concurrency channels coordinate mutex serializes · not parallelism · small interface sharp · zero value useful · `any` untyped tame it · gofmt settles bikeshed · tiny copy beats dep hairball syscall/cgo build tags isolate · cgo not Go · `unsafe` no contract · clarity beats wit · reflection stay cold path · errors values inspect wrap once · architecture name docs users · panic stays in `main` / hard startup.

## Uber style distill ([guide](https://github.com/uber-go/guide/blob/master/style.md), caveman compress)

Rare `*Iface` · `var _ I = (*T)(nil)` at export boundary · defer unlock pairs · chan buffer zero or one usually · slice/map copy exported API boundaries · typed errors `%w` chain handle once · assert comma-ok · goroutine bounded ctx/waitgroup · no zombie `init()` · globals inject not mutate · exits from `main` only · strconv hot paths · structs field-named literals · table tests sub `t.Run`.

---

Canon links above beat bullet memory when tradeoff unclear.
