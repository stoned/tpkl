package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/extreaders"
)

// ReadersCmd returns a cobra command to run custom readers for Pkl.
func ReadersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "readers",
		Short: "Run external readers for Pkl",
		Long:  "Run external readers for Pkl",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			extreaders.Run()
		},
	}

	return cmd
}
