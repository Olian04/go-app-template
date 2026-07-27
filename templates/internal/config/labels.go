// [[ when (modeIs "cli" "cli-library" "http") ]]
package config
[[ if modeIs "http" ]]

import (
	"fmt"

	"github.com/prometheus/common/model"
)
[[ end ]]

// Labels attach to slog[[ if modeIs "http" ]] and Prometheus const labels on metrics[[ end ]].
type Labels map[string]string

func (l Labels) WithDefaults() Labels {
	if l == nil {
		return Labels{}
	}
	out := make(Labels, len(l))
	for k, v := range l {
		out[k] = v
	}
	return out
}

func (l Labels) Validate() error {
[[ if modeIs "http" ]]
	for k := range l {
		if !model.LegacyValidation.IsValidLabelName(k) {
			return fmt.Errorf("invalid labels key %q: must satisfy legacy Prometheus label name rules", k)
		}
	}
[[ end ]]
	return nil
}
