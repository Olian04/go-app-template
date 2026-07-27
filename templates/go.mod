module [[.ModulePath]]

go [[.GoVersion]]

tool (
	github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	golang.org/x/vuln/cmd/govulncheck
)
[[- if modeIs "cli" "cli-library" "http" ]]

require (
[[- if modeIs "http" ]]
	github.com/prometheus/client_golang v1.23.2
[[- end ]]
	github.com/urfave/cli/v3 v3.8.0
	gopkg.in/yaml.v3 v3.0.1
)
[[- end ]]
