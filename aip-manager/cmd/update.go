package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"aip-manager/pkg/update"
	"github.com/spf13/cobra"
)

var (
	keyPath     string
	password    string
	sourceDir   string
	versionFile string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Assembles EDR files into a ZIP, then encrypts and signs as a .pkg file",
	Run: func(cmd *cobra.Command, args []string) {
		if password == "" {
			password = os.Getenv("ENCRYPT_PASSWORD")
		}
		if keyPath == "" {
			keyPath = os.Getenv("SIGNING_KEY_PATH")
		}
		if password == "" || keyPath == "" {
			log.Fatal("Error: Password and Signing Key path must be provided via flags or .env")
		}

		tempZipFile := filepath.Join(os.TempDir(), fmt.Sprintf("edr_update_%d.zip", time.Now().UnixNano()))
		defer os.Remove(tempZipFile)

		fmt.Println("Assembling layout into temporary zip...")
		err := update.BuildUpdateZip(sourceDir, versionFile, tempZipFile)
		if err != nil {
			log.Fatalf("Failed to build ZIP layout: %v", err)
		}

		outputPkg := fmt.Sprintf("agent_update_%d.pkg", time.Now().Unix())

		fmt.Printf("Encrypting, signing, and building %s...\n", outputPkg)
		err = update.BuildPackage(tempZipFile, keyPath, password, outputPkg)
		if err != nil {
			log.Fatalf("Update packaging failed: %v", err)
		}

		fmt.Println("Update package built successfully!")
	},
}

func init() {
	updateCmd.Flags().StringVarP(&sourceDir, "source", "s", "installer_src/file", "Path to the parent directory containing Common and Packed")
	updateCmd.Flags().StringVarP(&versionFile, "version", "v", "version.json", "Path to version.json")
	
	updateCmd.Flags().StringVarP(&keyPath, "key", "k", "", "Path to RSA Private Key PEM")
	updateCmd.Flags().StringVarP(&password, "password", "P", "", "Encryption Password")
	
	rootCmd.AddCommand(updateCmd)
}