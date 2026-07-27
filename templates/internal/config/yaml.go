// [[ when (modeIs "cli" "cli-library" "http") ]]
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// loadYAML decodes path into cfg. Empty path is a no-op (defaults / later sources apply).
func loadYAML(cfg *Config, path string) error {
	if path == "" {
		return nil
	}
	// #nosec G304 -- config path is explicit operator input, not derived from request data.
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var decoded Config
	if err := dec.Decode(&decoded); err != nil {
		return fmt.Errorf("parse config yaml: %w", err)
	}
	*cfg = decoded
	return nil
}
