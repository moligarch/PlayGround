// pkg/utils/report.go
// Minimal reporting + checksum helpers used by build/prepare pipelines.
// The build command will aggregate one BuildJobReport per config and
// write a single JSON report plus a checksums.txt file.

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BuildStepTiming carries per-job timings.
type BuildStepTiming struct {
	GenerateAIP time.Duration `json:"generate_aip"`
	BuildMSI    time.Duration `json:"build_msi"`
	Total       time.Duration `json:"total"`
}

// BuildJobReport captures the outcome for one config.
type BuildJobReport struct {
	ConfigPath     string            `json:"config_path"`
	NotePath       string            `json:"note_path,omitempty"`
	BuildTypeLabel string            `json:"build_type_label"` // e.g., basic, ai_packed
	ProductVersion string            `json:"product_version"`  // MSI ProductVersion used
	PackageBase    string            `json:"package_base"`     // e.g., installer_1.0.130.0_ai_packed
	AIPPath        string            `json:"aip_path,omitempty"`
	MSIPath        string            `json:"msi_path,omitempty"`
	Versions       map[string]string `json:"versions,omitempty"` // from version_includes.yaml (optional)
	Status         string            `json:"status"`             // "success" | "error" | "dry-run"
	Error          string            `json:"error,omitempty"`
	Timings        BuildStepTiming   `json:"timings"`
}

// BuildReport is the top-level artifact for a full run.
type BuildReport struct {
	GeneratedAt time.Time        `json:"generated_at"`
	HostWD      string           `json:"host_workdir"`
	Jobs        []BuildJobReport `json:"jobs"`
}

// WriteBuildReport writes a JSON report to path. Parent dirs are created if needed.
func WriteBuildReport(path string, r BuildReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

// FileChecksum contains the SHA256 checksum for one file.
type FileChecksum struct {
	Path     string `json:"path"`
	Algo     string `json:"algo"`     // e.g., "sha256"
	Checksum string `json:"checksum"` // hex
}

// ComputeSHA256 returns the hex-encoded SHA256 of the file at path.
func ComputeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer _closeNoErr(f)

	var h hash.Hash = sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WriteChecksums writes a text file (common format) with lines:
//
//	<sha256>  <relative-or-absolute-path>
//
// If any path does not exist, it is skipped.
func WriteChecksums(path string, filePaths []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create checksums dir: %w", err)
	}

	var sb stringsBuilder // tiny local builder avoiding fmt on tight loops
	for _, p := range filePaths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err != nil {
			// Skip missing files silently; the caller decides which artifacts should exist.
			continue
		}
		sum, err := ComputeSHA256(p)
		if err != nil {
			return fmt.Errorf("checksum %s: %w", p, err)
		}
		sb.WriteString(sum)
		sb.WriteString("  ")
		sb.WriteString(p)
		sb.WriteByte('\n')
	}

	return os.WriteFile(path, sb.Bytes(), 0o644)
}

// InferMSIPath returns "<dir>/<packageBase>.msi" given the AIP path + package base.
func InferMSIPath(aipPath, packageBase string) string {
	dir := filepath.Dir(aipPath)
	return filepath.Join(dir, packageBase+".msi")
}

// --- tiny strings builder (avoids importing bytes just for one use) ---

type stringsBuilder struct {
	b []byte
}

func (s *stringsBuilder) WriteString(str string) { s.b = append(s.b, str...) }
func (s *stringsBuilder) WriteByte(c byte)       { s.b = append(s.b, c) }
func (s *stringsBuilder) Bytes() []byte          { return s.b }

// --- internal helpers ---

func _closeNoErr(c io.Closer) {
	_ = c.Close()
}
