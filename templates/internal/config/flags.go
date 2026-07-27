// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

// FlagOverrides are applied last (highest precedence). Nil fields leave values unchanged.
type FlagOverrides struct {
	HTTPListenAddr *string
[[ if modeIs "http" ]]
	MetricsEnabled    *bool
	MetricsListenAddr *string
	MetricPrefix      *string
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	LoggingLevel  *string
	LoggingFormat *string
	LoggingStream *string
[[ end ]]
}

func applyFlags(cfg *Config, f FlagOverrides) {
	if f.HTTPListenAddr != nil {
		cfg.HTTP.ListenAddr = *f.HTTPListenAddr
	}
[[ if modeIs "http" ]]
	if f.MetricsListenAddr != nil {
		cfg.Metrics.ListenAddr = *f.MetricsListenAddr
	}
	if f.MetricPrefix != nil {
		cfg.Metrics.MetricPrefix = *f.MetricPrefix
	}
	if f.MetricsEnabled != nil {
		v := *f.MetricsEnabled
		cfg.Metrics.Enabled = &v
	}
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	if f.LoggingLevel != nil {
		cfg.Logging.Level = *f.LoggingLevel
	}
	if f.LoggingFormat != nil {
		cfg.Logging.Format = *f.LoggingFormat
	}
	if f.LoggingStream != nil {
		cfg.Logging.Stream = *f.LoggingStream
	}
[[ end ]]
}
