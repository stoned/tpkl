package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/internal/enumarg"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/tasks"
)

// GetListRunner returns a runner for the 'list' command.
func GetListRunner() *ListRunner {
	runner := &ListRunner{}
	runner.format = tasks.FormatEnumArg()

	command := &cobra.Command{
		Use:               "list",
		Short:             "List tasks",
		Long:              "List tpkl tasks defined in a Pkl module",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Run:               runner.Run,
	}

	addEnvFlag(command, &runner.env)
	addModuleFlag(command, &runner.module)
	addPropertyFlag(command, &runner.properties)
	addVerboseFlag(command, &runner.verbose)

	command.Flags().VarP(runner.format,
		"output", "o",
		"set output format. Supported formats: "+strings.Join(runner.format.Allowed, ", "))

	err := command.RegisterFlagCompletionFunc("output",
		cobra.FixedCompletions(runner.format.Allowed, cobra.ShellCompDirectiveDefault))
	if err != nil {
		logger := log.Builder(command.Name(), 0)
		logger.Fatal().Err(err).Send()
	}

	runner.command = command

	return runner
}

// ListCmd returns a Cobra command for the 'list' command.
func ListCmd() *cobra.Command {
	return GetListRunner().command
}

// ListRunner is a context for the 'list' command.
type ListRunner struct {
	command    *cobra.Command
	env        []string
	format     *enumarg.EnumArg
	module     string
	properties []string
	verbose    int
}

// Run runs the 'list' command.
func (r *ListRunner) Run(cmd *cobra.Command, _ []string) {
	ctx, logger := log.ContextWithLogger(cmd.Context(), "list", r.verbose)

	err := tasks.List(ctx, os.Stdout, r.format.String(),
		tasks.WithModule(r.module),
		tasks.WithEnv(r.env),
		tasks.WithProperties(r.properties))
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
