package cmd

import (
	"fmt"
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
		RunE:              runner.Run,
	}

	runner.command = command

	addEvalFlags(runner)

	command.Flags().VarP(runner.format,
		"output", "o",
		"Set output format. Supported formats: "+strings.Join(runner.format.Allowed, ", "))

	command.Flags().BoolVarP(&runner.long, "long", "l", false,
		`List tasks with documentation summary (same as output format "summary")`)

	err := command.RegisterFlagCompletionFunc("output",
		cobra.FixedCompletions(runner.format.Allowed, cobra.ShellCompDirectiveDefault))
	if err != nil {
		logger := log.Builder(command.Name(), 0)
		logger.Fatal().Err(err).Send()
	}

	return runner
}

// ListCmd returns a Cobra command for the 'list' command.
func ListCmd() *cobra.Command {
	return GetListRunner().command
}

// ListRunner is a context for the 'list' command.
type ListRunner struct {
	CommandRunner
	EvalRunner
	VerboseRunner

	format *enumarg.EnumArg
	long   bool
}

// Run runs the 'list' command.
func (r *ListRunner) Run(cmd *cobra.Command, _ []string) error {
	ctx, logger := log.ContextWithLogger(cmd.Context(), "list", r.verbose)

	flags := cmd.Flags()
	if flags.Changed("long") && flags.Changed("output") {
		return fmt.Errorf("%w: '%s' and '%s'", ErrMutuallyExclusiveFlags,
			flags.Lookup("long").Name, flags.Lookup("output").Name)
	}

	if r.long {
		err := r.format.Set("summary")
		if err != nil {
			return fmt.Errorf("error setting output format: %w", err)
		}
	}

	err := tasks.List(ctx, os.Stdout, r.format.String(),
		tasks.WithModule(r.module),
		tasks.WithEnv(r.env),
		tasks.WithProperties(r.properties),
		tasks.WithWorkingDir(r.workingDir))
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	return nil
}
