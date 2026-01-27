// pkg/enigma/runner.go
// ----------------------------------------------------------------------------
// Thin runner for Enigma Protector in profile mode.
// Executes: enigma64.exe -qe <profile.enigma64>
// No output printing here; callers (cmd) own user-facing logs.
// ----------------------------------------------------------------------------

package enigma

import (
	"aip-manager/pkg/utils"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// FindExe tries to locate enigma64.exe when Options.ExePath is empty.
//
// Precedence:
//  1. ENIGMA_EXE environment variable
//  2. Common Windows install locations (64/32-bit)
func FindExe() (string, error) {
	if p := os.Getenv("ENIGMA_EXE"); strings.TrimSpace(p) != "" {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\The Enigma Protector\enigma64.exe`,
			`C:\Program Files (x86)\The Enigma Protector\enigma64.exe`,
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}
	return "", errors.New("cannot find enigma64.exe: set ENIGMA_EXE or pass ExePath")
}

// RunWithOpts executes Enigma Protector with the provided options and returns
// combined output and the duration. On error, the message includes the last N
// lines of output if TailOnError > 0.
func RunWithOpts(opts Options) (string, time.Duration, error) {
	start := time.Now()

	if err := opts.Validate(); err != nil {
		return "", 0, err
	}

	exe := strings.TrimSpace(opts.ExePath)
	var err error
	if exe == "" {
		exe, err = FindExe()
		if err != nil {
			return "", 0, err
		}
	}

	// Context with timeout if requested.
	ctx := context.Background()
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, exe, abs(opts.ProfilePath))
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}

	out, runErr := cmd.CombinedOutput()
	dur := time.Since(start)
	outStr := string(out)

	if runErr != nil {
		// Timeout?
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			if opts.TailOnError > 0 {
				return outStr, dur, fmt.Errorf("enigma timed out after %s\n--- last %d lines ---\n%s",
					opts.Timeout, opts.TailOnError, utils.TailLines(outStr, opts.TailOnError))
			}
			return outStr, dur, fmt.Errorf("enigma timed out after %s", opts.Timeout)
		}
		if opts.TailOnError > 0 {
			return outStr, dur, fmt.Errorf("enigma failed: %v\n--- last %d lines ---\n%s",
				runErr, opts.TailOnError, utils.TailLines(outStr, opts.TailOnError))
		}
		return outStr, dur, fmt.Errorf("enigma failed: %v", runErr)
	}

	return outStr, dur, nil
}
