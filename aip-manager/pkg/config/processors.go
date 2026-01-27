// pkg/config/processors.go
package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// GenerateModulesPlaintext traverses the yaml.Node to create an INI string,
// preserving the original key order from the YAML file.
func (c *Config) GenerateModulesPlaintext() string {
	if c.Modules.Kind != yaml.MappingNode {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("[Modules]\n")

	// The Content of a MappingNode is a slice of [key, value, key, value, ...]
	for i := 0; i < len(c.Modules.Content); i += 2 {
		keyNode := c.Modules.Content[i]
		valueNode := c.Modules.Content[i+1]
		builder.WriteString(fmt.Sprintf("%s=%s\n", keyNode.Value, valueNode.Value))
	}
	return builder.String()
}

// GenerateSTSPlaintext traverses the nested yaml.Node structure for STS to create
// an INI string, preserving the original section and key order.
func (c *Config) GenerateSTSPlaintext() string {
	if c.STS.Kind != yaml.MappingNode {
		return ""
	}
	var builder strings.Builder

	// Iterate over sections (e.g., "general", "profiling")
	for i := 0; i < len(c.STS.Content); i += 2 {
		sectionKeyNode := c.STS.Content[i]
		sectionValueNode := c.STS.Content[i+1]

		builder.WriteString(fmt.Sprintf("[%s]\n", sectionKeyNode.Value))

		if sectionValueNode.Kind == yaml.MappingNode {
			// Iterate over key-value pairs within the section
			for j := 0; j < len(sectionValueNode.Content); j += 2 {
				itemKeyNode := sectionValueNode.Content[j]
				itemValueNode := sectionValueNode.Content[j+1]
				builder.WriteString(fmt.Sprintf("%s=%s\n", itemKeyNode.Value, itemValueNode.Value))
			}
		}
	}
	return builder.String()
}
