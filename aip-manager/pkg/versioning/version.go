// pkg/versioning/version.go
// Lightweight semantic version manager for the MSI build.
// - Format: MAJOR.MINOR.PATCH.BUILD (all non-negative integers)
// - Read/write from YAML (repo-local), e.g.:
//
//    major: 1
//    minor: 0
//    patch: 0
//    build: 0
//
// - Helpers to parse/format, validate, compare, and bump parts.
// - No ProductCode manipulation here — keeping ProductCode fixed is handled
//   simply by not changing it in the .aip (we only set ProductVersion).
//
// Typical usage:
//   v, _ := versioning.Load("aip-manager/version.yaml")
//   v.BumpPatch()              // or BumpBuild / BumpMinor / BumpMajor
//   _ = versioning.Save("aip-manager/version.yaml", v)
//   _ = doc.SetProductVersion(v.String())  // via aip.Document
//
// NOTE: We intentionally keep this package "pure" (no references to aip/cli).

package versioning

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version is a 4-part version: MAJOR.MINOR.PATCH.BUILD.
type Version struct {
	Major int `yaml:"major"`
	Minor int `yaml:"minor"`
	Patch int `yaml:"patch"`
	Build int `yaml:"build"`
}

// String returns "MAJOR.MINOR.PATCH.BUILD".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// Parse converts "a.b.c.d" into a Version.
func Parse(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return Version{}, fmt.Errorf("invalid version %q: expected 4 numeric parts", s)
	}
	nums := make([]int, 4)
	for i := 0; i < 4; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid numeric part %q in %q", parts[i], s)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Build: nums[3]}, nil
}

// Validate ensures all parts are >= 0.
func (v Version) Validate() error {
	if v.Major < 0 || v.Minor < 0 || v.Patch < 0 || v.Build < 0 {
		return errors.New("version parts must be non-negative")
	}
	return nil
}

// Compare returns -1/0/1 if v is < == > other (lexicographic on parts).
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	if v.Build != other.Build {
		if v.Build < other.Build {
			return -1
		}
		return 1
	}
	return 0
}

// Bump helpers. Each higher bump resets lower parts.
func (v *Version) BumpMajor() { v.Major++; v.Minor = 0; v.Patch = 0; v.Build = 0 }
func (v *Version) BumpMinor() { v.Minor++; v.Patch = 0; v.Build = 0 }
func (v *Version) BumpPatch() { v.Patch++; v.Build = 0 }
func (v *Version) BumpBuild() { v.Build++ }

// WithBuild returns a copy with a new Build value (>=0).
func (v Version) WithBuild(b int) Version {
	v.Build = b
	return v
}

// Load reads a YAML file from path into a Version.
// If the file doesn't exist, returns an error.
func Load(path string) (Version, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Version{}, err
	}
	var v Version
	if err := yaml.Unmarshal(raw, &v); err != nil {
		return Version{}, err
	}
	return v, v.Validate()
}

// Save writes the Version to path in YAML format (overwrites).
func Save(path string, v Version) error {
	if err := v.Validate(); err != nil {
		return err
	}
	out, err := yaml.Marshal(&v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// DefaultSearchPaths returns the preferred locations to look for version.yaml.
// We check under "aip-manager/version.yaml" first (to mirror your includes),
// then fall back to repo root "version.yaml".
func DefaultSearchPaths() []string {
	return []string{
		"aip-manager/version.yaml",
		"version.yaml",
	}
}

// FindAndLoad tries DefaultSearchPaths and returns the first version found.
func FindAndLoad() (Version, string, error) {
	for _, p := range DefaultSearchPaths() {
		if _, err := os.Stat(p); err == nil {
			v, err := Load(p)
			return v, p, err
		}
	}
	return Version{}, "", fmt.Errorf("version.yaml not found in any of: %v", DefaultSearchPaths())
}
