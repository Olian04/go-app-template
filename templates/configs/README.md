<!-- [[ when (modeIs "cli" "cli-library" "http") ]] -->
# Configuration

Optional layered sources. **Precedence**: **YAML → ENV → flags** (later overrides earlier).

| Layer | Notes |
| ----- | ----- |
| YAML  | File via `Options.Path` / `APP_CONFIG_FILE` at the process entrypoint. Example: `configs/config.example.yaml`. |
| ENV   | [[ if modeIs "http" ]]`APP_HTTP_LISTEN_ADDR`, `APP_METRICS_*`, [[ end ]]`APP_LOGGING_*`. |
| Flags | `FlagOverrides` on `Options.Flags` (nil field = leave unchanged). |

Sections present in this mode: `labels`[[ if modeIs "http" ]], `http`, `metrics`[[ end ]], `logging`.
[[ if modeIs "http" ]]
HTTP timeouts, `shutdown_timeout`, `max_header_bytes`, and `max_body_bytes` are YAML-only
(no ENV key); see `HTTPSection` in `internal/config/http.go` for defaults.
[[ end ]]

With no layers enabled, `Load` returns validated defaults only.
