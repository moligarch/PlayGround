// pkg/enigma/types.go
// ----------------------------------------------------------------------------
// Public types for the Enigma Protector integration (profile mode).
// Kept intentionally small: options + validation.
// ----------------------------------------------------------------------------

package enigma

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options configures a profile-mode Enigma Protector run.
//
// Only the profile-run scenario is supported (enigma64.exe -qe <profile>).
// The CLI executable path is optional; we auto-detect it when empty.
type Options struct {
	// ExePath: full path to enigma64.exe (optional, we auto-detect if empty).
	ExePath string

	// ProfilePath: path to the .enigma64 profile (required).
	ProfilePath string

	// Timeout bounds the process execution. 0 = no timeout.
	Timeout time.Duration

	// TailOnError: if > 0, append the last N lines of stdout/stderr to errors.
	TailOnError int

	// WorkingDir, if non-empty, is set as the process working directory.
	WorkingDir string
}

// Validate performs light sanity checks on fields.
func (o *Options) Validate() error {
	if o == nil {
		return errors.New("nil Options")
	}
	if strings.TrimSpace(o.ProfilePath) == "" {
		return errors.New("ProfilePath is required (.enigma64)")
	}
	if _, err := os.Stat(o.ProfilePath); err != nil {
		return fmt.Errorf("profile not found: %s (%w)", o.ProfilePath, err)
	}
	if o.WorkingDir != "" {
		if st, err := os.Stat(o.WorkingDir); err != nil || !st.IsDir() {
			return fmt.Errorf("working dir not found or not a dir: %s", o.WorkingDir)
		}
	}
	return nil
}

// abs returns an absolute path when possible; falls back to the input.
func abs(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
