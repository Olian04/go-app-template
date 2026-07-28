// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

// FlagOverrides are applied last (highest precedence). Nil fields leave values unchanged.
type FlagOverrides struct {
[[ if modeIs "http" ]]
	HTTPListenAddr    *string
	MetricsEnabled    *bool
	MetricsListenAddr *string
	MetricPrefix      *string
[[ end ]]
	LoggingLevel  *string
	LoggingFormat *string
	LoggingStream *string
}

func applyFlags(cfg *Config, f FlagOverrides) {
[[ if modeIs "http" ]]
	if f.HTTPListenAddr != nil {
		cfg.HTTP.ListenAddr = *f.HTTPListenAddr
	}
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
	if f.LoggingLevel != nil {
		cfg.Logging.Level = *f.LoggingLevel
	}
	if f.LoggingFormat != nil {
		cfg.Logging.Format = *f.LoggingFormat
	}
	if f.LoggingStream != nil {
		cfg.Logging.Stream = *f.LoggingStream
	}
}
