// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "installer",
	Short: "A CLI tool to manage and automate Advanced Installer project files (.aip). and building MSI",
	Long: `installer is a command-line application designed to automate
the configuration and generation of Advanced Installer project files.

It allows you to create different build variants from a single template
by dynamically updating registry keys, file paths, and other properties
based on a simple YAML configuration.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
