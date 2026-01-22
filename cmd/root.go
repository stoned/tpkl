package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// RootCmd returns a root cobra command for tpkl.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tpkl",
		Short:   "Tasks and tools for Pkl",
		Long:    "Run tasks defined in a Pkl module and readers for pkl command",
		Version: version(),
	}

	cmd.AddCommand(
		CatCmd(),
		DirCmd(),
		EvalCmd(),
		ListCmd(),
		ReadersCmd(),
		RunCmd(),
		VersionCmd(),
	)

	return cmd
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version
	}

	return "dev"
}
