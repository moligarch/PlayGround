// pkg/config/loader_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// findNodeValue is a helper function to find the value of a key in a mapping node.
func findNodeValue(node *yaml.Node, key string) (string, bool) {
	if node.Kind != yaml.MappingNode {
		return "", false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value, true
		}
	}
	return "", false
}

// findSectionNode is a helper to find a whole section (which is also a node).
func findSectionNode(node *yaml.Node, key string) (*yaml.Node, bool) {
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], true
		}
	}
	return nil, false
}

func TestLoadConfig(t *testing.T) {
	// 1. Setup: Generate sample config file.
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "inline.yaml")
	inline := []byte(`
buildType: "normal"
note_path: "configs/notes/note.core.json"
modules:
  AI_remove_quarantine_file: true
  Enable_Dac: false
sts:
  general:
    event_processor_workers: 3
    last_commit: "test-commit-hash"
`)

	if err := os.WriteFile(cfgPath, inline, 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	// 2. Act: Load the config file.
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 3. Assert: Check specific values to ensure they were parsed correctly.
	// We use the helper functions to traverse the yaml.Node structure.
	if cfg.BuildType != "normal" {
		t.Errorf("Expected buildType to be 'unpacked', but got '%s'", cfg.BuildType)
	}

	// Check a value in the 'modules' section
	aiRemoveVal, ok := findNodeValue(&cfg.Modules, "AI_remove_quarantine_file")
	if !ok || aiRemoveVal != "true" {
		t.Errorf("Expected modules.AI_remove_quarantine_file to be 'true', got '%s'", aiRemoveVal)
	}

	// Check a nested value in the 'sts' section
	generalSection, ok := findSectionNode(&cfg.STS, "general")
	if !ok {
		t.Fatal("Expected to find 'general' section in STS config")
	}

	workersVal, ok := findNodeValue(generalSection, "event_processor_workers")
	if !ok || workersVal != "3" {
		t.Errorf("Expected sts.general.event_processor_workers to be '3', but got '%s'", workersVal)
	}

	if got, want := cfg.NotePath, "configs/notes/note.core.json"; got != want {
		t.Fatalf("NotePath mismatch: got %q want %q", got, want)
	}
}
