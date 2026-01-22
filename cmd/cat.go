package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/modules"
)

// CatCmd returns a cobra command to print an embedded Pkl module.
func CatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat <module>",
		Short: "Print an embedded Pkl module",
		Long:  "Print a tpkl embedded Pkl module",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			modules := modules.Modules()
			if mod, ok := modules[args[0]]; ok {
				fmt.Print(mod)
			} else {
				fmt.Fprintf(os.Stderr, "Error: embedded module `%s` does not exist\n", args[0])
				os.Exit(1)
			}
		},
	}

	return cmd
}
