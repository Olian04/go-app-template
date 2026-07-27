// [[ when (modeIs "cli" "cli-library" "http") ]]
// Package config holds root config loading and subsection types (`labels`, `http`[[ if modeIs "http" ]], `metrics`[[ end ]][[ if modeIs "cli" "cli-library" "http" ]], `logging`[[ end ]]).
//
// Load precedence when multiple sources are enabled: YAML → ENV → flags (later wins).
package config

// Config is the aggregated application configuration.
type Config struct {
	Labels Labels      `yaml:"labels,omitempty"`
	HTTP   HTTPSection `yaml:"http,omitempty"`
[[ if modeIs "http" ]]
	Metrics MetricsSection `yaml:"metrics,omitempty"`
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	Logging LoggingSection `yaml:"logging,omitempty"`
[[ end ]]
}

func (c Config) WithDefaults() Config {
	c.Labels = c.Labels.WithDefaults()
	c.HTTP = c.HTTP.WithDefaults()
[[ if modeIs "http" ]]
	c.Metrics = c.Metrics.WithDefaults()
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	c.Logging = c.Logging.WithDefaults()
[[ end ]]
	return c
}

func (c Config) Validate() error {
	if err := c.Labels.Validate(); err != nil {
		return err
	}
	if err := c.HTTP.Validate(); err != nil {
		return err
	}
[[ if modeIs "http" ]]
	if err := c.Metrics.Validate(); err != nil {
		return err
	}
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	return c.Logging.Validate()
[[ else ]]
	return nil
[[ end ]]
}
[[ if modeIs "http" ]]

func (c Config) MetricsEnabled() bool {
	if c.Metrics.Enabled == nil {
		return true
	}
	return *c.Metrics.Enabled
}
[[ end ]]
