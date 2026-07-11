package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/tasks"
)

// GetManRunner returns a runner for the 'run' command.
func GetManRunner() *ManRunner {
	runner := &ManRunner{}
	command := &cobra.Command{
		Use:               "man <task> [flags]",
		Short:             "Show task documentation",
		Long:              "Show task documentation",
		Args:              cobra.ExactArgs(1),
		Run:               runner.Run,
		ValidArgsFunction: validArgsRun,
	}

	runner.command = command

	addEvalFlags(runner)

	command.Flags().BoolVarP(&runner.nopager, "no-pager", "P", false,
		`Do not pipe documentation into a pager`)

	return runner
}

// ManCmd returns a Cobra command for the 'man' command.
func ManCmd() *cobra.Command {
	return GetManRunner().command
}

// ManRunner is a context for the 'run' command.
type ManRunner struct {
	CommandRunner
	EvalRunner
	VerboseRunner

	nopager bool
}

// Run runs the 'man' command.
func (r *ManRunner) Run(cmd *cobra.Command, args []string) {
	ctx, logger := log.ContextWithLogger(cmd.Context(), "man", r.verbose)

	err := tasks.Man(ctx, os.Stdout, args[0],
		tasks.WithNoPager(r.nopager),
		tasks.WithEnv(r.env),
		tasks.WithModule(r.module),
		tasks.WithProperties(r.properties))
	if err != nil {
		log.AsFatal(logger, err.Error())

		e := (&tasks.CmdError{})
		if errors.As(err, &e) {
			os.Exit(e.ExitCode)
		}

		os.Exit(1)
	}
}
