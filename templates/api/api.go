// [[ when (modeIs "http") ]]
// Package api exposes this service's OpenAPI contract.
//
// The spec is embedded rather than read from disk so a deployed binary always
// serves the contract it was built from, with no runtime file dependency.
package api

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// SpecYAML is the OpenAPI document as authored.
//
//go:embed openapi.yaml
var SpecYAML []byte

// SpecJSON converts the document to JSON for tooling that will not read YAML.
//
// Returns an error for a malformed spec so callers can fail at startup instead
// of serving a broken contract.
func SpecJSON() ([]byte, error) {
	var doc any
	if err := yaml.Unmarshal(SpecYAML, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi.yaml: %w", err)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("encode openapi json: %w", err)
	}
	return out, nil
}
