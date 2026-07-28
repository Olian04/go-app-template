// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

// Options selects call-site inputs for enabled config sources.
type Options struct {
	// Path to YAML config file. Empty skips the YAML layer (defaults apply).
	Path string

	// Flags applied last (highest precedence). Nil fields leave values unchanged.
	Flags FlagOverrides
}

// Load builds Config from enabled sources.
//
// Precedence: YAML → ENV → flags. Later sources override earlier ones.
func Load(opts Options) (Config, error) {
	var cfg Config
	if err := loadYAML(&cfg, opts.Path); err != nil {
		return Config{}, err
	}
	cfg = cfg.WithDefaults()
	if err := applyEnv(&cfg); err != nil {
		return Config{}, err
	}
	applyFlags(&cfg, opts.Flags)
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
