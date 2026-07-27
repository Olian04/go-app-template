# Agent context: [[ .ModuleBasename ]]

App template — use with repo `README.md`[[ if modeIs "cli" "cli-library" "http" ]] and [`configs/config.example.yaml`](../configs/config.example.yaml)[[ end ]]. Mode: `[[ .Mode ]]`.

## Layout

| Path | Role |
| --- | --- |
[[ if modeIs "http" ]]
| `cmd/[[ .ServiceName ]]` | Entry: load config, `logging.Setup`, signal context, `app.Run`. |
| `internal/app` | Wire Prometheus registry, router, dual listeners. |
| `internal/domain/[[ .ServiceName ]]` | Domain logic only. |
| `internal/transport/http` | Routes + middleware (request logs + metrics increment). |
| `internal/transport/metricshttp` | `/metrics` server lifecycle. |
[[ end ]]
[[ if modeIs "cli" "cli-library" ]]
| `cmd/[[ .CliName ]]` | CLI entrypoint. |
[[ end ]]
[[ if modeIs "library" "cli-library" ]]
| `pkg/[[ .LibName ]]` | Public library API. |
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
| `internal/config` | Root config `Load`; section structs (defaults/`WithDefaults`/`Validate`). |
| `internal/observability/logging` | slog setup. |
| `configs` | YAML contract examples. |
[[ end ]]
[[ if modeIs "http" ]]
| `internal/observability/metrics` | Prometheus registry + handler. |
[[ end ]]
| `test/unit/...` | Unit tests beside mirrored paths. |

## Dependency direction

`cmd` → `internal/app` → domain, transport[[ if modeIs "cli" "cli-library" "http" ]], observability[[ end ]][[ if modeIs "cli" "cli-library" "http" ]], aggregated `internal/config`[[ end ]]. Domain must not import app, transports, observability.

[[ if modeIs "cli" "cli-library" "http" ]]
## Configuration

`APP_CONFIG_FILE` optional; unset uses defaults matching example file shapes.

Section structs carry `Default*` constants in-package, `WithDefaults()`, `Validate()`. Root `config.Config.Validate()` delegates down.
[[ end ]]

## Mode notes

[[ if modeIs "cli" "cli-library" "http" ]]
- **Logging**: use `internal/observability/logging` for slog setup.
[[ else ]]
- **Logging**: standard library `slog` defaults (no `observability/logging` package).
[[ end ]]
[[ if modeIs "http" ]]
- **Metrics**: Prometheus registry + `/metrics` listener via `metricshttp`.
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
