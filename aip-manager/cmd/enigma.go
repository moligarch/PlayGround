// cmd/enigma.go
// ----------------------------------------------------------------------------
// Enigma Protector (profile mode) runner for CI/local use.
//
// What this command does:
//   1) Rewrites the provided .enigma64 profile so its input/output files point
//      to the *current workspace* Unpacked/Packed directories.
//   2) Saves the rewritten profile at: build/<yyyy-MM-dd>/enigma_profile.enigma64
//   3) Executes: enigma64.exe -qe <rewritten profile>, with timeout + tail-on-error.
//
// Flags:
//   --profile, -p   : source .enigma64 profile (required)
//   --in, -i        : Unpacked dir (default: installer_src/files/Unpacked)
//   --out, -o       : Packed dir   (default: installer_src/files/Packed)
//   --exe           : path to enigma64.exe (optional; ENIGMA_EXE env also supported)
//   --timeout       : e.g., 10m, 120s; 0 = no timeout
//   --tail          : on error, include last N lines of Enigma output
// ----------------------------------------------------------------------------

package cmd

import (
	"aip-manager/pkg/enigma"
	"aip-manager/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	enigmaProfile string        // source .enigma64 (required)
	enigmaExe     string        // optional explicit path to enigma64.exe
	enigmaTimeout time.Duration // e.g. 5m, 120s; 0 = no timeout
	enigmaTail    int           // on error, include last N lines of output

	// Workspace locations:
	unpackedDir string // --in / -i
	packedDir   string // --out / -o
)

var enigmaCmd = &cobra.Command{
	Use:   "enigma",
	Short: "Rewrite an Enigma profile for this workspace and run it (profile mode).",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()

		if strings.TrimSpace(enigmaProfile) == "" {
			fmt.Fprintln(os.Stderr, "Error: --profile is required (path to .enigma64)")
			os.Exit(1)
		}
		// Validate dirs (Packed is created if missing)
		if st, err := os.Stat(unpackedDir); err != nil || !st.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: Unpacked dir not found or not a dir: %s\n", unpackedDir)
			os.Exit(1)
		}
		if err := os.MkdirAll(packedDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot create Packed dir: %v\n", err)
			os.Exit(1)
		}

		// 1) Rewrite profile to point at actual job paths
		fmt.Println("🧩 Rewriting Enigma profile paths...")
		profT := utils.StartTimer("Rewriting profile")
		rewrittenPath, err := enigma.GenerateProfileWithMappedPaths(enigmaProfile, unpackedDir, packedDir)
		utils.EndTimer(profT)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: rewrite profile failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("   • Source profile:    %s\n", enigmaProfile)
		fmt.Printf("   • Unpacked dir:      %s\n", absOr(unpackedDir))
		fmt.Printf("   • Packed dir:        %s\n", absOr(packedDir))
		fmt.Printf("   • Rewritten profile: %s\n", rewrittenPath)

		// 2) Run Enigma Protector with the rewritten profile
		fmt.Println("🔐 Running Enigma Protector (profile mode)...")
		runT := utils.StartTimer("Enigma")
		out, dur, err := enigma.RunWithOpts(enigma.Options{
			ExePath:     enigmaExe,
			ProfilePath: rewrittenPath,
			Timeout:     enigmaTimeout,
			TailOnError: enigmaTail,
		})
		utils.EndTimer(runT)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if enigmaTail == 0 && len(out) > 0 {
				fmt.Println("Hint: re-run with --tail 80 to include the last lines of Enigma output.")
			}
			os.Exit(1)
		}
		fmt.Printf("✅ Enigma completed in %s\n", dur.Truncate(time.Millisecond))
		// Uncomment to see raw console output:
		// fmt.Println("--- Enigma Output ---")
		// fmt.Println(out)
		// fmt.Println("---------------------")
	},
}

func init() {
	rootCmd.AddCommand(enigmaCmd)
	enigmaCmd.Flags().StringVarP(&enigmaProfile, "profile", "p", "", "Path to Enigma Protector profile (.enigma64) [required]")
	enigmaCmd.Flags().StringVar(&enigmaExe, "exe", "", "Path to enigma64.exe (optional; ENIGMA_EXE env also supported)")
	enigmaCmd.Flags().DurationVar(&enigmaTimeout, "timeout", 0, "Timeout for Enigma run (e.g. 5m, 120s). 0 = no timeout")
	enigmaCmd.Flags().IntVar(&enigmaTail, "tail", 0, "On error, include last N lines of Enigma output")
	enigmaCmd.Flags().StringVarP(&unpackedDir, "in", "i", filepath.Join("installer_src", "files", "Unpacked"), "Directory containing source files to pack")
	enigmaCmd.Flags().StringVarP(&packedDir, "out", "o", filepath.Join("installer_src", "files", "Packed"), "Directory to place packed outputs")
}

func absOr(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
