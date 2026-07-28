// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

import (
[[ if modeIs "http" ]]
	"fmt"
[[ end ]]
	"os"
[[ if modeIs "http" ]]
	"strconv"
[[ end ]]
	"strings"
)

// ENV keys (applied after YAML, before flags):
//
[[ if modeIs "http" ]]
//	APP_HTTP_LISTEN_ADDR
//	APP_METRICS_ENABLED
//	APP_METRICS_LISTEN_ADDR
//	APP_METRICS_METRIC_PREFIX
[[ end ]]
//	APP_LOGGING_LEVEL
//	APP_LOGGING_FORMAT
//	APP_LOGGING_STREAM
[[ if modeIs "http" ]]
//
// HTTP timeouts and size limits are YAML/flag-only; see HTTPSection.
[[ end ]]
func applyEnv(cfg *Config) error {
[[ if modeIs "http" ]]
	if v, ok := lookupEnv("APP_HTTP_LISTEN_ADDR"); ok {
		cfg.HTTP.ListenAddr = v
	}
	if v, ok := lookupEnv("APP_METRICS_LISTEN_ADDR"); ok {
		cfg.Metrics.ListenAddr = v
	}
	if v, ok := lookupEnv("APP_METRICS_METRIC_PREFIX"); ok {
		cfg.Metrics.MetricPrefix = v
	}
	if v, ok := lookupEnv("APP_METRICS_ENABLED"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("APP_METRICS_ENABLED: %w", err)
		}
		cfg.Metrics.Enabled = &b
	}
[[ end ]]
	if v, ok := lookupEnv("APP_LOGGING_LEVEL"); ok {
		cfg.Logging.Level = v
	}
	if v, ok := lookupEnv("APP_LOGGING_FORMAT"); ok {
		cfg.Logging.Format = v
	}
	if v, ok := lookupEnv("APP_LOGGING_STREAM"); ok {
		cfg.Logging.Stream = v
	}
	return nil
}

func lookupEnv(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	return v, true
}
