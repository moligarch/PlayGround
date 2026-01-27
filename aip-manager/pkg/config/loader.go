// pkg/config/loader.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses a YAML configuration file from the given path.
// It returns a pointer to a Config struct populated with the file's data.
func LoadConfig(path string) (*Config, error) {
	// Read the entire file into a byte slice.
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Create an empty Config struct to hold the parsed data.
	var cfg Config
	// Unmarshal the YAML data into the struct. The yaml.v3 library
	// automatically maps the YAML keys to the struct fields.
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
