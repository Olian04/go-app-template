// [[ when (modeIs "cli" "cli-library" "http") ]]
package config
[[ if modeIs "http" ]]

import "time"
[[ end ]]

const (
	DefaultLoggingLevel  = "info"
	DefaultLoggingFormat = "json"
	DefaultLoggingStream = "stdout"
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
