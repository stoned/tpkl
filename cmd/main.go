// Package cmd provides tpkl's CLI framework via Cobra
package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Main is tpkl main cmd entrypoint.
func Main() {
	rootCmd := RootCmd()

	// if at least one argument is provided check if it is a valid command
	// or else assume it is a task to be run
	if len(os.Args) > 1 {
		var (
			cmd *cobra.Command
			err error
		)

		rootCmd.InitDefaultHelpCmd()
		rootCmd.InitDefaultCompletionCmd(os.Args[1:]...)

		if rootCmd.TraverseChildren {
			cmd, _, err = rootCmd.Traverse(os.Args[1:])
		} else {
			cmd, _, err = rootCmd.Find(os.Args[1:])
		}

		if err != nil && cmd.Use == rootCmd.Use && strings.HasPrefix(err.Error(), "unknown command") {
			args := append([]string{"run"}, os.Args[1:]...)
			rootCmd.SetArgs(args)
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
