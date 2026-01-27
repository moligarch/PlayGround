// cmd/generate_aip.go
// Prepare (generate) a .aip without building the MSI.
// Uses concise utils; prints minimal header and timings.

package cmd

import (
	"aip-manager/pkg/config"
	"aip-manager/pkg/utils"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var generateAipCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Generate a build-ready .aip file without building the MSI.",
	Run: func(cmd *cobra.Command, args []string) {
		total := utils.StartTimer("🚀 Starting aip preparation ...")
		_ = godotenv.Load()

		configPath, _ := cmd.Flags().GetString("config")
		templatePath, _ := cmd.Flags().GetString("template")

		// Header info
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config: %v\n", err)
			os.Exit(1)
		}
		buildLabel := utils.ComputeBuildTypeLabel(cfg.BuildType, configPath)

		fmt.Printf("\n==> [1/1] build started\n")
		fmt.Printf(`
==============================================
   +  Config:     %s
   +  Note:       %s
   +  Build:      %s
==============================================

`, configPath, cfg.NotePath, buildLabel)

		// Generate AIP (concise)
		res, genDur, err := utils.GenerateAIPConcise(templatePath, configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate AIP: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("   ⏱  Generate AIP: %s\n", genDur.Truncate(time.Millisecond))
		fmt.Printf("🎉 AIP file generated successfully at: %s\n", res.AIPPath)

		utils.EndTimer(total) // overall total
	},
}

func init() {
	rootCmd.AddCommand(generateAipCmd)
	generateAipCmd.Flags().StringP("config", "c", "", "Path to the config YAML file")
	generateAipCmd.Flags().StringP("template", "t", "templates/base.aip", "Path to the template .aip file")
	_ = generateAipCmd.MarkFlagRequired("config")
}
