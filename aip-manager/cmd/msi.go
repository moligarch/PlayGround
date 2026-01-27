// cmd/msi.go
// Build an MSI from an existing .aip file.
// Uses concise utils; prints step time and overall total.

package cmd

import (
	"aip-manager/pkg/utils"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var msiCmd = &cobra.Command{
	Use:   "msi",
	Short: "Build an MSI from an existing .aip file.",
	Run: func(cmd *cobra.Command, args []string) {
		total := utils.StartTimer("🚀 Building MSI from AIP ...")
		_ = godotenv.Load()

		aipPath, _ := cmd.Flags().GetString("aip")
		if aipPath == "" {
			fmt.Fprintln(os.Stderr, "Error: --aip is required")
			os.Exit(1)
		}

		fmt.Printf("\n==> build from AIP\n")
		fmt.Printf("   AIP: %s\n", aipPath)

		out, buildDur, err := utils.BuildFromAIPConcise(aipPath, "")
		if err != nil {
			// On error, print AI output to help debugging.
			fmt.Println("--- Advanced Installer Output ---")
			fmt.Println(out)
			fmt.Println("-------------------------------")
			fmt.Fprintf(os.Stderr, "MSI build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("   ⏱  Build MSI: %s\n", buildDur.Truncate(time.Millisecond))
		utils.EndTimer(total) // overall total
	},
}

func init() {
	rootCmd.AddCommand(msiCmd)
	msiCmd.Flags().String("aip", "", "Path to an existing .aip file")
	_ = msiCmd.MarkFlagRequired("aip")
}
