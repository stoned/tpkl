package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCmd returns a cobra command to print tpkl's version.
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version for tpkl",
		Long:  "Version for tpkl",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("tpkl version %s\n", version())
		},
	}

	return cmd
}
