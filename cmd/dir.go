package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/modules"
)

// DirCmd returns a cobra command to list tpkl's embedded Pkl modules.
func DirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dir",
		Short: "List embedded Pkl modules",
		Long:  "List tpkl embedded Pkl modules",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			for modname := range modules.Modules() {
				fmt.Println(modname)
			}
		},
	}

	return cmd
}
