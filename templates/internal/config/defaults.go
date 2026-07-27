// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

const (
	DefaultHTTPListenAddr = ":8080"
[[ if modeIs "http" ]]
	DefaultMetricsListenAddr = ":9090"
	DefaultMetricPrefix      = "[[ .ModuleBasename ]]"
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	DefaultLoggingLevel  = "info"
	DefaultLoggingFormat = "json"
	DefaultLoggingStream = "stdout"
[[ end ]]
)
[[ if modeIs "http" ]]

const DefaultMetricsEnabled = true
[[ end ]]
