package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/log"
)

// ErrMutuallyExclusiveFlags signals an error about mutually exclusive flags.
var ErrMutuallyExclusiveFlags = errors.New("mutually exclusive flags specified")

// CommandRunner is a context for commands.
type CommandRunner struct {
	command *cobra.Command
}

func (r *CommandRunner) commandAddr() *cobra.Command {
	return r.command
}

// EvalRunner is a context for commands which run a Pkl evaluation.
type EvalRunner struct {
	env        []string
	module     string
	properties []string
}

func (r *EvalRunner) envAddr() *[]string {
	return &r.env
}

func (r *EvalRunner) moduleAddr() *string {
	return &r.module
}

func (r *EvalRunner) propertiesAddr() *[]string {
	return &r.properties
}

// VerboseRunner is a  context for commands which take a verbose option.
type VerboseRunner struct {
	verbose int
}

func (r *VerboseRunner) verboseAddr() *int {
	return &r.verbose
}

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

func addEvalFlags(runner evalRunner) {
	cmd := runner.commandAddr()
	addEnvFlag(cmd, runner.envAddr())
	addModuleFlag(cmd, runner.moduleAddr())
	addPropertyFlag(cmd, runner.propertiesAddr())
	addVerboseFlag(cmd, runner.verboseAddr())
}

type evalRunner interface {
	commandAddr() *cobra.Command
	envAddr() *[]string
	moduleAddr() *string
	propertiesAddr() *[]string
	verboseAddr() *int
}
