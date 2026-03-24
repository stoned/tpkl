// Package cmd provides tpkl's CLI framework via Cobra
package cmd

import (
	"context"
	"os"
)

// Main is tpkl main cmd entrypoint.
func Main() {
	rootCmd := RootCmd()
	rootCmd.SetContext(context.Background())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
