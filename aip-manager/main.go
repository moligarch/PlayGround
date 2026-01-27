// main.go
package main

import (
	"aip-manager/cmd"
)

func main() {
	// This is the primary entry point for the application.
	// It calls the Execute function from the root command
	// defined in the cmd package.
	cmd.Execute()
}
