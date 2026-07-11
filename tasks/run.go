package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hmdsefi/gograph"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/stoned/tpkl/internal/expansion"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/modules/tpkl"
)

// RunOptions are options for Run().
type runOptions struct {
	evalOptions

	args    []string
	dryrun  bool
	timeout time.Duration
}

// RunOption is Run()'s options interface.
type RunOption interface {
	setRunOption(opts *runOptions)
}

// Set arguments Run()'s option.
func (o *argsOption) setRunOption(opts *runOptions) {
	opts.args = o.args
}

// Set dryRun Run()'s option.
func (o *dryrunOption) setRunOption(opts *runOptions) {
	opts.dryrun = o.dryrun
}

// Set env Run()'s option.
func (o *envOption) setRunOption(opts *runOptions) {
	opts.env = o.env
}

// Set module Run()'s option.
func (o *moduleOption) setRunOption(opts *runOptions) {
	opts.module = o.module
}

// Set properties Run()'s option.
func (o *propertiesOption) setRunOption(opts *runOptions) {
	opts.properties = o.properties
}

// Set timeout Run()'s option.
func (o *timeoutOption) setRunOption(opts *runOptions) {
	opts.timeout = o.timeout
}

// Run executes task from a Pkl module.
func Run(ctx context.Context, taskName string, options ...RunOption) error {
	var (
		err    error
		cancel context.CancelFunc
	)

	opts := &runOptions{}
	for _, opt := range options {
		opt.setRunOption(opts)
	}

	logger := log.FromContext(ctx).With().Str("task", taskName).Logger()
	ctx = logger.WithContext(ctx)

	opts.module, err = UseModule(ctx, opts.module, opts.workingDir)
	if err != nil {
		return fmt.Errorf("run task: %w", err)
	}

	if opts.timeout != 0 {
		ctx, cancel = context.WithTimeoutCause(ctx, opts.timeout, ErrTimeout)
		defer cancel()
	}

	frame := NewTopFrame(taskName, opts.module, opts.env, opts.args)

	tasks, err := ModuleTasks(ctx, opts.module, WithPklEnv(frame.EnvList()),
		WithPklProperties(opts.properties))
	if err != nil {
		return err
	}

	err = planTask(ctx, taskName, tasks)
	if err != nil {
		return err
	}

	termChannel, termWaitGroup := termHandler()

	err = runTask(ctx, taskName, tasks, frame, termChannel, termWaitGroup, opts)
	if context.Cause(ctx) != nil {
		err = fmt.Errorf("%w: %w", context.Cause(ctx), err)
	}

	return err
}

func termHandler() (chan any, *sync.WaitGroup) {
	termChannel := make(chan any)
	termWaitGroup := new(sync.WaitGroup)
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigChannel
		close(termChannel)
		termWaitGroup.Wait()

		os.Exit(128)
	}()

	return termChannel, termWaitGroup
}

// NewTopFrame create a toplevel frame for a task.
func NewTopFrame(taskName string, module string, env []string, args []string) *Frame {
	frame := NewFrame()

	for _, variable := range env {
		name, val, found := strings.Cut(variable, "=")
		if !found {
			val = ""
		}

		frame.SetVar(name, val)
	}

	frame.setPrefixedVar("ARGC", strconv.Itoa(len(os.Args)))

	for i, a := range os.Args {
		frame.setPrefixedVar("ARG_"+strconv.Itoa(i), a)
	}

	frame.setPrefixedVar("TASK_ARGC", strconv.Itoa(len(args)))

	for i, a := range args {
		frame.setPrefixedVar("TASK_ARG_"+strconv.Itoa(i), a)
	}

	frame.setPrefixedVar("MODULE", module)
	frame.setPrefixedVar("MODULEDIR", moduleDir(module))
	frame.setPrefixedVar("TASK", taskName)

	return frame
}

func newTaskFrame(taskName string, task tpkl.Task, enclosingFrame *Frame) *Frame {
	frame := NewEnclosedFrame(enclosingFrame)

	frame.SetVars(task.Env)

	if task.InheritEnv {
		frame.SetEnviron()
	}

	frame.setPrefixedVar("CURRENT_TASK", taskName)

	return frame
}

type planNode struct {
	task     tpkl.Task
	children []*planNode
}

// planTask returns an error if it determines the task can't be run.
// XXX use this function to also detect/warn about task file shadowing ?
func planTask(_ context.Context, start string, tasks Tasks) error {
	var plan func(string, string) (*planNode, error)

	taskExists := func(n string) bool {
		_, ok := tasks[n]

		return ok
	}

	if !taskExists(start) {
		return fmt.Errorf("%w: `%s`", ErrUnknownTask, start)
	}

	planGraph := gograph.New[string](gograph.Acyclic())

	plan = func(parent, name string) (*planNode, error) {
		task := tasks[name]
		node := &planNode{task: task}

		for _, cmd := range task.Cmds {
			if cmd.Task == nil {
				continue
			}

			if !taskExists(*cmd.Task) {
				return nil, fmt.Errorf("plan for task `%s`: %w: `%s`", start, ErrUnknownTask, *cmd.Task)
			}

			if parent != "" {
				fromVertex := gograph.NewVertex(parent)
				toVertex := gograph.NewVertex(name)

				if planGraph.GetEdge(fromVertex, toVertex) == nil {
					_, err := planGraph.AddEdge(fromVertex, toVertex)
					if err != nil {
						return nil, fmt.Errorf("plan task `%s`: %w: calling task `%s` from task `%s`: %w",
							start, ErrTaskCycle, name, parent, err)
					}
				}
			}

			child, err := plan(name, *cmd.Task)
			if err != nil {
				return node, err
			}

			if child != nil {
				node.children = append(node.children, child)
			}
		}

		return node, nil
	}

	// XXX drop the constructed (and returned) task's call tree
	// if we ended up doing nothing with it!
	_, err := plan("", start)

	return err
}

func runTask(ctx context.Context, taskName string,
	tasks Tasks, enclosingFrame *Frame,
	termChannel chan any, termWaitGroup *sync.WaitGroup,
	opts *runOptions,
) error {
	var cmdErr error

	logger := log.FromContext(ctx).With().Str("cur", taskName).Logger()
	ctx = logger.WithContext(ctx)

	if _, ok := tasks[taskName]; !ok {
		return fmt.Errorf("%w: `%s`", ErrUnknownTask, taskName)
	}

	task := tasks[taskName]
	frame := newTaskFrame(taskName, task, enclosingFrame)

	taskFiles, err := newTaskFiles(task, frame, termChannel, termWaitGroup)
	if taskFiles != nil {
		defer func() {
			_ = taskFiles.cleanup()
		}()
	}

	if err != nil {
		return err
	}

	expandTaskProperties(&task, frame.ExpandMapping())

	logger.Trace().Str("workdir", task.WorkingDir).Send()

	for cmdIdx, cmd := range task.Cmds {
		switch {
		case cmd.Task != nil:
			logger.Info().Str("call", *cmd.Task).Send()
			cmdErr = runTask(ctx, *cmd.Task, tasks, frame, termChannel, termWaitGroup, opts)

		case cmd.EmbeddedShell:
			log.Shell(ctx, cmd.Cmd)

			scriptName := fmt.Sprintf("%s[%d]", taskName, cmdIdx)
			cmdErr = runShell(ctx, scriptName, cmd.Cmd, task.WorkingDir, frame.EnvList(), opts.dryrun)

		default:
			log.Cmd(ctx, cmd.Cmd)

			cmdErr = runCmd(ctx, cmd.Cmd, task.WorkingDir, frame.EnvList(), opts.dryrun)
		}

		if cmdErr != nil {
			if cmd.MustSucceed {
				logger.Err(cmdErr).Msg("command failed")

				return cmdErr
			}

			logger.Info().Err(cmdErr).Msg("ignoring failed command")
		}
	}

	return nil
}

// runCmd runs an arbitrary command.
func runCmd(ctx context.Context, command []string, dir string, environ []string, dryrun bool) error {
	if dryrun {
		for idx, word := range command {
			quoted, err := syntax.Quote(word, syntax.LangBash)
			if err != nil {
				fmt.Print(word)
			} else {
				fmt.Print(quoted)
			}

			if idx != len(command)-1 {
				fmt.Print(" ")
			}
		}

		fmt.Print("\n")

		return nil
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = environ

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ee := &exec.ExitError{}

	err := cmd.Run()
	if errors.As(err, &ee) {
		ec := ee.ExitCode()

		return NewCmdError(ec, err)
	}

	if err != nil {
		return NewCmdError(1, err)
	}

	return nil
}

// runShell runs an arbitrary shell command or script with a mvdan.cc shell interpreter, so called
// "embedded shell" in tpkl. cf. https://github.com/mvdan/sh
func runShell(ctx context.Context, taskName string, command []string, dir string, environ []string, dryrun bool) error {
	parser, err := syntax.NewParser().Parse(strings.NewReader(command[0]), taskName)
	if err != nil {
		return NewCmdError(1, err)
	}

	runner, err := interp.New(
		interp.Params(command[1:]...),
		interp.Dir(dir),
		interp.Env(expand.ListEnviron(environ...)),
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
	)
	if err != nil {
		return NewCmdError(1, err)
	}

	if dryrun {
		fmt.Print(command[0], "\n")

		return nil
	}

	err = runner.Run(ctx, parser)
	if err != nil {
		var es interp.ExitStatus
		if errors.As(err, &es) {
			return NewCmdError(int(es), err)
		}

		return NewCmdError(1, err)
	}

	return nil
}

type taskFile struct {
	Path    string
	Varname *string
}
type taskFiles struct {
	Dir   string
	Files map[string]taskFile
}

func newTaskFiles(task tpkl.Task, frame *Frame,
	termChannel chan any, termWaitGroup *sync.WaitGroup,
) (*taskFiles, error) {
	tfiles := taskFiles{Files: make(map[string]taskFile)}

	if len(task.Files) == 0 {
		return &tfiles, nil
	}

	dir, err := os.MkdirTemp(os.TempDir(), "tpkl_taskfiles_*")
	if err != nil {
		return nil, fmt.Errorf("%w: error creating temporary directory: %w", ErrTaskFile, err)
	}

	termWaitGroup.Add(1)

	tfiles.Dir = dir

	go func() {
		defer termWaitGroup.Done()

		<-termChannel

		_ = os.RemoveAll(tfiles.Dir)
	}()

	for key, file := range task.Files {
		var filename string

		if file.Filename != nil {
			filename = *file.Filename
		} else {
			filename = key
		}

		path := filepath.Join(tfiles.Dir, filename)

		tmpfile, err := os.Create(path) // #nosec G304
		if err != nil {
			return &tfiles, fmt.Errorf("%w: creating temporary file: %q: %w", ErrTaskFile, path, err)
		}

		tfiles.Files[key] = taskFile{Path: path, Varname: file.Varname}

		_, err = tmpfile.WriteString(file.Content)
		if err != nil {
			return &tfiles, fmt.Errorf("%w: writing to temporary file: %q: %w", ErrTaskFile, path, err)
		}

		err = tmpfile.Close()
		if err != nil {
			return &tfiles, fmt.Errorf("%w: closing temporary file: %q: %w", ErrTaskFile, path, err)
		}
	}

	err = frame.setTaskFilesVars(&tfiles)
	if err != nil {
		return &tfiles, err
	}

	return &tfiles, nil
}

func (tf *taskFiles) cleanup() error {
	if len(tf.Dir) == 0 {
		return nil
	}

	err := os.RemoveAll(tf.Dir)
	if err != nil {
		return fmt.Errorf("%w: removing temporary directory: %s: %w", ErrTaskFile, tf.Dir, err)
	}

	return nil
}

// Set variables relative to task files in a frame.
func (f *Frame) setTaskFilesVars(files *taskFiles) error {
	var (
		err                            error
		fileIndex, enclosingFilesCount int
	)

	// Get TPKL_FILES_COUNT from enclosing frame if any
	if f.enclosing != nil {
		enclosingVars := f.enclosing.Merge()
		name := prefixedVarName(filesCountVarNameSuffix)

		if value, ok := enclosingVars[name]; ok {
			enclosingFilesCount, err = strconv.Atoi(value)
			if err == nil {
				fileIndex = enclosingFilesCount
			} else {
				return fmt.Errorf("%w: converting environment variable `%s` value to integer: %w", ErrTaskFile, name, err)
			}
		}
	}

	// Set TPKL_FILES_COUNT for this frame.
	f.setPrefixedVar(filesCountVarNameSuffix, strconv.Itoa(enclosingFilesCount+len(files.Files)))

	if len(files.Dir) != 0 {
		f.setPrefixedVar("FILES_DIR", files.Dir)
	}

	for key, file := range files.Files {
		f.setPrefixedVar("FILE_"+key, file.Path)
		f.setPrefixedVar("FILES_KEY_"+strconv.Itoa(fileIndex), key)
		f.setPrefixedVar("FILES_PATH_"+strconv.Itoa(fileIndex), file.Path)

		if file.Varname != nil {
			f.SetVar(*file.Varname, file.Path)
		}

		fileIndex++
	}

	return nil
}

func expandTaskProperties(task *tpkl.Task, mapping func(string) string) {
	task.WorkingDir = expansion.Expand(task.WorkingDir, mapping)

	for _, cmd := range task.Cmds {
		for i, word := range cmd.Cmd {
			cmd.Cmd[i] = expansion.Expand(word, mapping)
		}
	}
}

func (f *Frame) setPrefixedVar(name string, value string) {
	f.SetVar(prefixedVarName(name), value)
}

func prefixedVarName(name string) string {
	return identifierPrefix + name
}
