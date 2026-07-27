<!-- [[ when (modeIs "cli" "cli-library" "http") ]] -->
# Configuration

Optional layered sources. **Precedence**: **YAML → ENV → flags** (later overrides earlier).

| Layer | Notes |
| ----- | ----- |
| YAML  | File via `Options.Path` / `APP_CONFIG_FILE` at the process entrypoint. Example: `configs/config.example.yaml`. |
| ENV   | `APP_HTTP_LISTEN_ADDR`[[ if modeIs "http" ]], `APP_METRICS_*`[[ end ]][[ if modeIs "cli" "cli-library" "http" ]], `APP_LOGGING_*`[[ end ]]. |
| Flags | `FlagOverrides` on `Options.Flags` (nil field = leave unchanged). |

With no layers enabled, `Load` returns validated defaults only.
