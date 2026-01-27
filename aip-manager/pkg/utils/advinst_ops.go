//go:build windows

// pkg/utils/find_advinst.go
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// FindAdvancedInstallerPath searches the Windows Registry for the installation path.
// It checks both 64-bit and 32-bit (Wow6432Node) registry locations and looks for
// either "InstallRoot" or "Path" as the value name.
func FindAdvancedInstallerPath() (string, error) {
	pathsToTry := []string{
		`SOFTWARE\Caphyon\Advanced Installer`,             // Standard 64-bit location
		`SOFTWARE\Wow6432Node\Caphyon\Advanced Installer`, // 32-bit location on 64-bit Windows
	}

	for _, path := range pathsToTry {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if err != nil {
			// If this path doesn't exist, just continue to the next one.
			continue
		}
		defer key.Close()

		// Try to read "InstallRoot" first, as you specified.
		installPath, _, err := key.GetStringValue("InstallRoot")
		if err != nil {
			continue
		}

		if installPath != "" {
			// If we found a valid path, construct the full executable path and return.
			fullPath := filepath.Join(installPath, "bin", "x86", "AdvancedInstaller.com")
			return fullPath, nil
		}
	}

	// If the loop completes without finding anything, return the final error.
	return "", fmt.Errorf("advanced installer registry key not found. Is it installed?")
}

// CleanAdvInstallerCaches scans 'dir' for folders ending with "-cache" and
// removes them. It returns the number of removed folders and their paths.
// If the directory doesn't exist, it returns (0, nil, nil).
func CleanAdvInstallerCaches(dir string) (int, []string, error) {
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil, nil
		}
		return 0, nil, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !st.IsDir() {
		return 0, nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil, fmt.Errorf("readdir %s: %w", dir, err)
	}

	removed := 0
	var removedPaths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, "-cache") {
			continue
		}
		p := filepath.Join(dir, name)
		if rmErr := os.RemoveAll(p); rmErr != nil {
			return removed, removedPaths, fmt.Errorf("remove %s: %w", p, rmErr)
		}
		removed++
		removedPaths = append(removedPaths, p)
	}
	return removed, removedPaths, nil
}
