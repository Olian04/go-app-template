// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

// Options selects call-site inputs for enabled config sources.
type Options struct {
[[ if modeIs "cli" "cli-library" "http" ]]
	// Path to YAML config file. Empty skips the YAML layer (defaults apply).
	Path string
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	// Flags applied last (highest precedence). Nil fields leave values unchanged.
	Flags FlagOverrides
[[ end ]]
}

// Load builds Config from enabled sources.
//
// Precedence (enabled sources only): YAML → ENV → flags.
// Later sources override earlier ones. Disabled sources are omitted entirely.
func Load(opts Options) (Config, error) {
	var cfg Config
[[ if modeIs "cli" "cli-library" "http" ]]
	if err := loadYAML(&cfg, opts.Path); err != nil {
		return Config{}, err
	}
[[ end ]]
	cfg = cfg.WithDefaults()
[[ if modeIs "cli" "cli-library" "http" ]]
	if err := applyEnv(&cfg); err != nil {
		return Config{}, err
	}
[[ end ]]
[[ if modeIs "cli" "cli-library" "http" ]]
	applyFlags(&cfg, opts.Flags)
[[ end ]]
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
