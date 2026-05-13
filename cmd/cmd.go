package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/log"
)

// ErrMutuallyExclusiveFlags signals an error about mutually exclusive flags.
var ErrMutuallyExclusiveFlags = errors.New("mutually exclusive flags specified")

func addEnvFlag(cmd *cobra.Command, variable *[]string) {
	cmd.Flags().StringArrayVarP(variable, "env-var", "e", nil, "Set environment variable `name[=value]` (repeatable)")

	err := cmd.RegisterFlagCompletionFunc("env-var", cobra.NoFileCompletions)
	if err != nil {
		logger := log.Builder(cmd.Name(), 0)
		logger.Fatal().Err(err).Send()
	}
}

func addModuleFlag(cmd *cobra.Command, variable *string) {
	cmd.Flags().StringVarP(variable, "module", "m", "", "Set Pkl module path or URI")

	err := cmd.RegisterFlagCompletionFunc("module",
		cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault))
	if err != nil {
		logger := log.Builder(cmd.Name(), 0)
		logger.Fatal().Err(err).Send()
	}
}

func addPropertyFlag(cmd *cobra.Command, variable *[]string) {
	cmd.Flags().StringArrayVarP(variable, "property", "p", nil, "Set external property `name[=value]` (repeatable)")

	err := cmd.RegisterFlagCompletionFunc("property", cobra.NoFileCompletions)
	if err != nil {
		logger := log.Builder(cmd.Name(), 0)
		logger.Fatal().Err(err).Send()
	}
}

func addVerboseFlag(cmd *cobra.Command, variable *int) {
	cmd.Flags().CountVarP(variable, "v", "v", "Set log `level[=+1]` verbosity (repeatable)")

	err := cmd.RegisterFlagCompletionFunc("v", cobra.NoFileCompletions)
	if err != nil {
		logger := log.Builder(cmd.Name(), 0)
		logger.Fatal().Err(err).Send()
	}
}
