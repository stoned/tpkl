package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/tasks"
)

// GetRunRunner returns a runner for the 'run' command.
func GetRunRunner() *RunRunner {
	runner := &RunRunner{}
	command := &cobra.Command{
		Use:   "run <task> [flags] [-- args]...",
		Short: "Run task (default)",
		Long:  "Run tpkl task from a Pkl module",
		Args:  cobra.MinimumNArgs(1),
		Run:   runner.Run,
	}

	addEnvFlag(command, &runner.env)
	addModuleFlag(command, &runner.module)
	addPropertyFlag(command, &runner.properties)
	addVerboseFlag(command, &runner.verbose)
	runner.timeout = command.Flags().DurationP("timeout", "t", 0,
		"Duration after which task execution will be timed out")

	runner.command = command

	return runner
}

// RunCmd returns a Cobra command for the 'run' command.
func RunCmd() *cobra.Command {
	return GetRunRunner().command
}

// RunRunner is a context for the 'run' command.
type RunRunner struct {
	command    *cobra.Command
	env        []string
	module     string
	properties []string
	timeout    *time.Duration
	verbose    int
}

// Run runs the 'run' command.
func (r *RunRunner) Run(_ *cobra.Command, args []string) {
	logger := log.New("run", r.verbose)
	ctx := logger.WithContext(context.Background())

	err := tasks.Run(ctx, args[0],
		tasks.WithArgs(args[1:]),
		tasks.WithEnv(r.env),
		tasks.WithModule(r.module),
		tasks.WithProperties(r.properties),
		tasks.WithVerbosity(r.verbose), // XXX not needed anymore?
		tasks.WithTimeout(r.timeout))
	if err != nil {
		log.AsFatal(&logger, err.Error())

		e := (&tasks.CmdError{})
		if errors.As(err, &e) {
			os.Exit(e.ExitCode)
		}

		os.Exit(1)
	}
}
