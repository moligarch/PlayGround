// pkg/config/processors_test.go
package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateModulesPlaintext(t *testing.T) {
	// The yaml.v3 library preserves key order when unmarshalling into a yaml.Node.
	// This test verifies that our generator function respects that order.
	yamlContent := `
modules:
  Enable_Dac: false
  AI_remove_quarantine_file: true
  Blacklist_remove_quarantine_file: true
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlContent), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// The order in the expected string MUST match the order in the yamlContent.
	expected := "[Modules]\n" +
		"Enable_Dac=false\n" +
		"AI_remove_quarantine_file=true\n" +
		"Blacklist_remove_quarantine_file=true\n"

	actual := cfg.GenerateModulesPlaintext()
	if actual != expected {
		t.Errorf("GenerateModulesPlaintext() output mismatch.\nExpected:\n%q\nGot:\n%q", expected, actual)
	}
}

func TestGenerateSTSPlaintext(t *testing.T) {
	// This test verifies order preservation for the more complex, nested STS config.
	yamlContent := `
sts:
  profiling:
    enable_profiling: 0
  general:
    hide_console: 0
    enable_self_defense: 1
    last_commit: "abc-123"
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlContent), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// The order of both sections and keys MUST match the yamlContent.
	expected := "[profiling]\n" +
		"enable_profiling=0\n" +
		"[general]\n" +
		"hide_console=0\n" +
		"enable_self_defense=1\n" +
		"last_commit=abc-123\n"

	actual := cfg.GenerateSTSPlaintext()
	if actual != expected {
		t.Errorf("GenerateSTSPlaintext() output mismatch.\nExpected:\n%q\nGot:\n%q", expected, actual)
	}
}
