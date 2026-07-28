// [[ when (modeIs "cli" "cli-library" "http") ]]
package config_test

import (
	"testing"
[[ if modeIs "http" ]]
	"time"
[[ end ]]

	"[[.ModulePath]]/internal/config"
)

// Defaults must survive Validate, otherwise a zero-config start fails.
func TestDefaultsValidate(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
}

func TestLoggingDefaults(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	if cfg.Logging.Level != config.DefaultLoggingLevel {
		t.Fatalf("level = %q, want %q", cfg.Logging.Level, config.DefaultLoggingLevel)
	}
	if cfg.Logging.Format != config.DefaultLoggingFormat {
		t.Fatalf("format = %q, want %q", cfg.Logging.Format, config.DefaultLoggingFormat)
	}
}

func TestLoggingRejectsBadLevel(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	cfg.Logging.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Fatal("want error for logging.level=verbose")
	}
}
[[ if modeIs "http" ]]

func TestHTTPDefaults(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	if cfg.HTTP.ListenAddr != config.DefaultHTTPListenAddr {
		t.Fatalf("listen_addr = %q, want %q", cfg.HTTP.ListenAddr, config.DefaultHTTPListenAddr)
	}
	// Every bound must be positive or the server would inherit "no limit".
	if cfg.HTTP.ReadTimeout <= 0 || cfg.HTTP.WriteTimeout <= 0 ||
		cfg.HTTP.IdleTimeout <= 0 || cfg.HTTP.ReadHeaderTimeout <= 0 ||
		cfg.HTTP.ShutdownTimeout <= 0 {
		t.Fatalf("non-positive timeout in %+v", cfg.HTTP)
	}
	if cfg.HTTP.MaxBodyBytes <= 0 || cfg.HTTP.MaxHeaderBytes <= 0 {
		t.Fatalf("non-positive size limit in %+v", cfg.HTTP)
	}
}

func TestHTTPRejectsNonPositiveTimeout(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	cfg.HTTP.ReadTimeout = -time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("want error for negative http.read_timeout")
	}
}

func TestHTTPRejectsEmptyListenAddr(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	cfg.HTTP.ListenAddr = "  "
	if err := cfg.Validate(); err == nil {
		t.Fatal("want error for blank http.listen_addr")
	}
}

func TestMetricsEnabledDefaultsTrue(t *testing.T) {
	if !(config.Config{}).WithDefaults().MetricsEnabled() {
		t.Fatal("metrics should default to enabled")
	}
}

func TestMetricsPrefixMustBeValid(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	cfg.Metrics.MetricPrefix = "bad-prefix"
	if err := cfg.Validate(); err == nil {
		t.Fatal("want error for hyphenated metric prefix")
	}
}

func TestLabelsMustBeValidPrometheusNames(t *testing.T) {
	cfg := config.Config{Labels: config.Labels{"not a label": "x"}}.WithDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("want error for invalid label name")
	}
}
[[ end ]]
