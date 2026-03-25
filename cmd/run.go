package cmd

import (
	"errors"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/tasks"
)

func validArgsRun(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	module, _ := cmd.Flags().GetString("module")
	env, _ := cmd.Flags().GetStringArray("env-var")
	properties, _ := cmd.Flags().GetStringArray("property")
	ctx := cmd.Context()

	module, err := tasks.UseModule(ctx, module, "") // XXX support working-dir
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	frame := tasks.NewTopFrame("", module, env, nil)

	tasks, err := tasks.ModuleTasks(ctx, module,
		tasks.WithPklEnv(frame.EnvList()),
		tasks.WithPklProperties(properties),
		tasks.WithPropertyListCommandRunning())
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	completions := make([]string, 0, len(tasks))

	names := slices.Sorted(maps.Keys(tasks))
	for _, task := range names {
		// XXX get task description
		completions = append(completions, cobra.CompletionWithDesc(task, "Task "+task))
	}

	return completions, cobra.ShellCompDirectiveDefault
}

func addRunFlags(runner *RunRunner) {
	cmd := runner.command
	addEnvFlag(cmd, &runner.env)
	addModuleFlag(cmd, &runner.module)
	addPropertyFlag(cmd, &runner.properties)
	addVerboseFlag(cmd, &runner.verbose)
	cmd.Flags().BoolVarP(&runner.dryrun, "dry-run", "n", false, "Do not execute tasks commands, print them")
	cmd.Flags().DurationVarP(&runner.timeout, "timeout", "t", 0,
		"Duration after which task execution will be timed out")

	err := cmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)
	if err != nil {
		logger := log.Builder("run", 0)
		logger.Fatal().Err(err).Send()
	}
}

// GetRunRunner returns a runner for the 'run' command.
func GetRunRunner() *RunRunner {
	runner := &RunRunner{}
	command := &cobra.Command{
		Use:               "run <task> [flags] [-- args]...",
		Short:             "Run task (default)",
		Long:              "Run tpkl task from a Pkl module",
		Args:              cobra.MinimumNArgs(1),
		Run:               runner.Run,
		ValidArgsFunction: validArgsRun,
	}

	runner.command = command

	addRunFlags(runner)

	return runner
}

// RunCmd returns a Cobra command for the 'run' command.
func RunCmd() *cobra.Command {
	return GetRunRunner().command
}

// RunRunner is a context for the 'run' command.
type RunRunner struct {
	command    *cobra.Command
	dryrun     bool
	env        []string
	module     string
	properties []string
	timeout    time.Duration
	verbose    int
}

// Run runs the 'run' command.
func (r *RunRunner) Run(cmd *cobra.Command, args []string) {
	ctx, logger := log.ContextWithLogger(cmd.Context(), "run", r.verbose)

	err := tasks.Run(ctx, args[0],
		tasks.WithArgs(args[1:]),
		tasks.WithDryrun(r.dryrun),
		tasks.WithEnv(r.env),
		tasks.WithModule(r.module),
		tasks.WithProperties(r.properties),
		tasks.WithTimeout(r.timeout))
	if err != nil {
		log.AsFatal(logger, err.Error())

		e := (&tasks.CmdError{})
		if errors.As(err, &e) {
			os.Exit(e.ExitCode)
		}

		os.Exit(1)
	}
}
