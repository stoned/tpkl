package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// RootCmd returns a root cobra command for tpkl.
func RootCmd() *cobra.Command {
	runner := &RunRunner{}
	command := &cobra.Command{
		Use:               "tpkl [flags]\n  tpkl <task> [flags] [-- args]...",
		Short:             "Tasks and tools for Pkl",
		Long:              "Run tasks defined in a Pkl module and readers for pkl command",
		Version:           version(),
		Args:              cobra.MinimumNArgs(1),
		Run:               runner.Run,
		ValidArgsFunction: validArgsRun,
	}

	runner.command = command

	addRunFlags(runner)

	command.AddCommand(
		CatCmd(),
		DirCmd(),
		EvalCmd(),
		ListCmd(),
		ManCmd(),
		ReadersCmd(),
		RunCmd(),
		VersionCmd(),
	)

	return command
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version
	}

	return "dev"
}
