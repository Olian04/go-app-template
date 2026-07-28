// [[ when (modeIs "cli" "cli-library" "http") ]]
// Package config holds root config loading and subsection types (`labels`[[ if modeIs "http" ]], `http`, `metrics`[[ end ]], `logging`).
//
// Load precedence when multiple sources are enabled: YAML → ENV → flags (later wins).
package config

// Config is the aggregated application configuration.
type Config struct {
	Labels Labels `yaml:"labels,omitempty"`
[[ if modeIs "http" ]]
	HTTP    HTTPSection    `yaml:"http,omitempty"`
	Metrics MetricsSection `yaml:"metrics,omitempty"`
[[ end ]]
	Logging LoggingSection `yaml:"logging,omitempty"`
}

func (c Config) WithDefaults() Config {
	c.Labels = c.Labels.WithDefaults()
[[ if modeIs "http" ]]
	c.HTTP = c.HTTP.WithDefaults()
	c.Metrics = c.Metrics.WithDefaults()
[[ end ]]
	c.Logging = c.Logging.WithDefaults()
	return c
}

func (c Config) Validate() error {
	if err := c.Labels.Validate(); err != nil {
		return err
	}
[[ if modeIs "http" ]]
	if err := c.HTTP.Validate(); err != nil {
		return err
	}
	if err := c.Metrics.Validate(); err != nil {
		return err
	}
[[ end ]]
	return c.Logging.Validate()
}
[[ if modeIs "http" ]]

func (c Config) MetricsEnabled() bool {
	if c.Metrics.Enabled == nil {
		return true
	}
	return *c.Metrics.Enabled
}
[[ end ]]
