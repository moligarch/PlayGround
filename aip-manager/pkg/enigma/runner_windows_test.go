//go:build windows

package enigma

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildFakeExe compiles a tiny Go program to an .exe used to simulate Enigma.
func buildFakeExe(t *testing.T, dir, name, src string) string {
	t.Helper()
	if !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}
	srcPath := filepath.Join(dir, name+".go")
	exePath := filepath.Join(dir, name)
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", exePath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
	return exePath
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// 1) Success path: prints args, sleeps ~2s, exit 0.
func TestEnigma_Success(t *testing.T) {
	tmp := t.TempDir()

	src := `package main
import (
	"fmt"
	"os"
	"time"
)
func main() {
	fmt.Println("[FAKE-ENIGMA] args:", os.Args[1:])
	time.Sleep(2 * time.Second)
}`
	exe := buildFakeExe(t, tmp, "fake_enigma_success", src)

	profile := filepath.Join(tmp, "profile.enigma64")
	writeFile(t, profile, "dummy")

	out, dur, err := RunWithOpts(Options{
		ExePath:     exe,
		ProfilePath: profile,
		Timeout:     10 * time.Second,
		TailOnError: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nOutput:\n%s", err, out)
	}
	if dur <= 0 {
		t.Fatalf("expected positive duration, got %v", dur)
	}
	if !strings.Contains(out, "[FAKE-ENIGMA] args:") {
		t.Fatalf("expected args line in output, got:\n%s", out)
	}
}

// 2) Timeout behavior: sleeps ~120s, timeout ~1.5s.
func TestEnigma_Timeout(t *testing.T) {
	tmp := t.TempDir()

	src := `package main
import (
	"fmt"
	"time"
)
func main() {
	fmt.Println("[FAKE-ENIGMA] long run...")
	time.Sleep(120 * time.Second)
}`
	exe := buildFakeExe(t, tmp, "fake_enigma_timeout", src)

	profile := filepath.Join(tmp, "profile.enigma64")
	writeFile(t, profile, "dummy")

	out, _, err := RunWithOpts(Options{
		ExePath:     exe,
		ProfilePath: profile,
		Timeout:     1500 * time.Millisecond,
		TailOnError: 10,
	})
	if err == nil {
		t.Fatalf("expected timeout error, got nil (out=%q)", out)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "timed out") {
		t.Fatalf("expected timeout message, got: %v", err)
	}
}

// 3) Tail-on-error: print several lines, exit non-zero; expect last 2 lines in error.
func TestEnigma_TailOnError(t *testing.T) {
	tmp := t.TempDir()

	src := `package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Println("[FAKE-ENIGMA] starting...")
	fmt.Println("step 1")
	fmt.Println("step 2")
	fmt.Println("step 3 (the end)")
	os.Exit(13)
}`
	exe := buildFakeExe(t, tmp, "fake_enigma_fail", src)

	profile := filepath.Join(tmp, "profile.enigma64")
	writeFile(t, profile, "dummy")

	out, _, err := RunWithOpts(Options{
		ExePath:     exe,
		ProfilePath: profile,
		TailOnError: 2,
	})
	if err == nil {
		t.Fatalf("expected error, got nil (out=%q)", out)
	}
	msg := err.Error()
	if !strings.Contains(msg, "step 2") || !strings.Contains(msg, "step 3 (the end)") {
		t.Fatalf("expected last 2 lines in error; got:\n%v\nOutput:\n%s", msg, out)
	}
}

// 4) Validation errors: missing/empty ProfilePath.
func TestEnigma_ValidationErrors(t *testing.T) {
	tmp := t.TempDir()

	src := `package main
func main() {}`
	exe := buildFakeExe(t, tmp, "fake_enigma_nop", src)

	// Case 1: empty ProfilePath
	if _, _, err := RunWithOpts(Options{ExePath: exe}); err == nil {
		t.Fatalf("expected error for missing profile path")
	}
	// Case 2: non-existent profile file
	missing := filepath.Join(tmp, "nope.enigma64")
	if _, _, err := RunWithOpts(Options{ExePath: exe, ProfilePath: missing}); err == nil {
		t.Fatalf("expected error for non-existent profile path")
	}
}
