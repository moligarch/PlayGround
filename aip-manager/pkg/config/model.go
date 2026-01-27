// pkg/config/model.go
package config

import "gopkg.in/yaml.v3"

// Config represents the structure of the main YAML configuration file.
//   - Generate INI strings from yaml.Node (order-preserving)
//   - Upstream code encrypts those plaintext INI strings the same way as before
//     when writing to the AIP
type Config struct {
	// Which binary tree to target when resolving paths (e.g., "normal" or "packed").
	BuildType string `yaml:"buildType"`

	// Path to the JSON note payload to embed (moved from CLI into config files).
	// Example:
	//   note_path: "configs/notes/note.basic.json"
	NotePath string `yaml:"note_path"`

	// Plaintext sections (order-preserving via yaml.Node).
	Modules yaml.Node `yaml:"modules,omitempty"`
	STS     yaml.Node `yaml:"sts,omitempty"`
}
