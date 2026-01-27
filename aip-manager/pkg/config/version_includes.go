// pkg/config/version_includes.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// VersionIncludes models the repo-level version_includes.yaml,
// which lists binaries/DLLs to read version info from at build time.
// Example YAML:
//   includes:
//     - key: cybercore
//       path: installer_src/files/Common/CyberCore.exe
//     - key: ui_version
//       path: installer_src/files/Common/AgentNotification/AgentNotification.exe
type VersionIncludes struct {
	Includes []struct {
		Key  string `yaml:"key"`
		Path string `yaml:"path"`
	} `yaml:"includes"`
}

// LoadVersionIncludes reads version_includes.yaml-like files.
func LoadVersionIncludes(path string) (*VersionIncludes, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var vi VersionIncludes
	if err := yaml.Unmarshal(data, &vi); err != nil {
		return nil, err
	}
	return &vi, nil
}
