// cmd/version.go
// Versioning CLI for managing version.yaml used to set ProductVersion in AIP.
// Subcommands:
//   - version show
//   - version set [a.b.c.d]  OR with flags: --major --minor --patch --build
//   - version bump (--major | --minor | --patch | --build)
//
// Notes:
// - The file is searched in preferred locations via versioning.FindAndLoad().
// - On `set`, if no file exists yet, it creates one at the first preferred path.

package cmd

import (
	"aip-manager/pkg/versioning"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show or change the MSI version (stored in version.yaml).",
}

var versionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current version and the path to version.yaml.",
	Run: func(cmd *cobra.Command, args []string) {
		v, path, err := versioning.FindAndLoad()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Version: %s\nFile:    %s\n", v.String(), path)
	},
}

var (
	setMajor int
	setMinor int
	setPatch int
	setBuild int
)

var versionSetCmd = &cobra.Command{
	Use:   "set [a.b.c.d]",
	Short: "Set the version explicitly (positional a.b.c.d or with flags).",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var v versioning.Version
		var targetPath string
		var err error

		// Try to load existing to preserve path if present.
		if cur, path, loadErr := versioning.FindAndLoad(); loadErr == nil {
			v = cur
			targetPath = path
		} else {
			// No existing file; choose the first preferred path.
			paths := versioning.DefaultSearchPaths()
			targetPath = paths[0]
		}

		if len(args) == 1 {
			// Parse from positional a.b.c.d
			v, err = versioning.Parse(strings.TrimSpace(args[0]))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Use flags if provided; if none provided, error.
			flagsProvided := false
			if cmd.Flags().Changed("major") {
				v.Major = setMajor
				flagsProvided = true
			}
			if cmd.Flags().Changed("minor") {
				v.Minor = setMinor
				flagsProvided = true
			}
			if cmd.Flags().Changed("patch") {
				v.Patch = setPatch
				flagsProvided = true
			}
			if cmd.Flags().Changed("build") {
				v.Build = setBuild
				flagsProvided = true
			}
			if !flagsProvided {
				fmt.Fprintln(os.Stderr, "Error: provide either a positional a.b.c.d or one/more of --major/--minor/--patch/--build")
				os.Exit(1)
			}
			if err := v.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Ensure directory exists for new file.
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory for %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		if err := versioning.Save(targetPath, v); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set version to %s (%s)\n", v.String(), targetPath)
	},
}

var (
	bumpMajor bool
	bumpMinor bool
	bumpPatch bool
	bumpBuild bool
)

var versionBumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump one part of the version (use exactly one of --major|--minor|--patch|--build).",
	Run: func(cmd *cobra.Command, args []string) {
		v, path, err := versioning.FindAndLoad()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		which, err := exactlyOneTrue(
			map[string]bool{
				"major": bumpMajor,
				"minor": bumpMinor,
				"patch": bumpPatch,
				"build": bumpBuild,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		switch which {
		case "major":
			v.BumpMajor()
		case "minor":
			v.BumpMinor()
		case "patch":
			v.BumpPatch()
		case "build":
			v.BumpBuild()
		default:
			fmt.Fprintln(os.Stderr, "Error: use exactly one of --major | --minor | --patch | --build")
			os.Exit(1)
		}

		if err := versioning.Save(path, v); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Bumped %s → %s (%s)\n", which, v.String(), path)
	},
}

func init() {
	// wire up hierarchy
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionShowCmd)
	versionCmd.AddCommand(versionSetCmd)
	versionCmd.AddCommand(versionBumpCmd)

	// flags: set
	versionSetCmd.Flags().IntVar(&setMajor, "major", 0, "Major version (non-negative)")
	versionSetCmd.Flags().IntVar(&setMinor, "minor", 0, "Minor version (non-negative)")
	versionSetCmd.Flags().IntVar(&setPatch, "patch", 0, "Patch version (non-negative)")
	versionSetCmd.Flags().IntVar(&setBuild, "build", 0, "Build number (non-negative)")

	// flags: bump (exactly one)
	versionBumpCmd.Flags().BoolVar(&bumpMajor, "major", false, "Bump MAJOR (resets minor, patch, build)")
	versionBumpCmd.Flags().BoolVar(&bumpMinor, "minor", false, "Bump MINOR (resets patch, build)")
	versionBumpCmd.Flags().BoolVar(&bumpPatch, "patch", false, "Bump PATCH (resets build)")
	versionBumpCmd.Flags().BoolVar(&bumpBuild, "build", false, "Bump BUILD")
}

// exactlyOneTrue returns the single key whose value is true, else error.
func exactlyOneTrue(m map[string]bool) (string, error) {
	var (
		found string
		count int
	)
	for k, v := range m {
		if v {
			found = k
			count++
		}
	}
	if count != 1 {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, "--"+k)
		}
		return "", errors.New("use exactly one of " + strings.Join(keys, "|"))
	}
	return found, nil
}
