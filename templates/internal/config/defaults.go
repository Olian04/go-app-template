// [[ when (modeIs "cli" "cli-library" "http") ]]
package config
[[ if modeIs "http" ]]

import "time"
[[ end ]]

const (
	DefaultLoggingLevel  = "info"
	DefaultLoggingFormat = "json"
[[ if modeIs "http" ]]
	// A service's logs are its primary output, so stdout is the conventional sink.
	DefaultLoggingStream = "stdout"
[[ else ]]
	// A CLI's stdout is its data channel: keep logs on stderr so results stay
	// pipeable. Override with logging.stream or APP_LOGGING_STREAM.
	DefaultLoggingStream = "stderr"
[[ end ]]
)
[[ if modeIs "http" ]]

const (
	DefaultHTTPListenAddr        = ":8080"
	DefaultHTTPReadTimeout       = 15 * time.Second
	DefaultHTTPReadHeaderTimeout = 5 * time.Second
	DefaultHTTPWriteTimeout      = 15 * time.Second
	DefaultHTTPIdleTimeout       = 60 * time.Second
	DefaultHTTPShutdownTimeout   = 10 * time.Second
	DefaultHTTPMaxHeaderBytes    = 1 << 20 // 1 MiB
	DefaultHTTPMaxBodyBytes      = 1 << 20 // 1 MiB
)

const (
	DefaultMetricsListenAddr = ":9090"
	DefaultMetricPrefix      = "[[ .MetricPrefix ]]"
	DefaultMetricsEnabled    = true
)
[[ end ]]
