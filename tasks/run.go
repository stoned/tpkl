package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

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
	args       []string
	env        []string
	module     string
	properties []string
	timeout    *time.Duration
	verbose    int
	workingDir string
}

// RunOption is Run()'s options interface.
type RunOption interface {
	setRunOption(o *runOptions)
}

// Set arguments Run()'s option.
func (a *argsOption) setRunOption(o *runOptions) {
	o.args = a.args
}

// Set env Run()'s option.
func (e *envOption) setRunOption(o *runOptions) {
	o.env = e.env
}

// Set module Run()'s option.
func (m *moduleOption) setRunOption(o *runOptions) {
	o.module = m.module
}

// Set properties Run()'s option.
func (p *propertiesOption) setRunOption(o *runOptions) {
	o.properties = p.properties
}

// Set timeout Run()'s option.
func (t *timeoutOption) setRunOption(o *runOptions) {
	o.timeout = t.timeout
}

// Set verbose Run()'s option.
func (v *verbosityOption) setRunOption(o *runOptions) {
	o.verbose = v.verbose
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

	opts.module, err = useModule(ctx, opts.module, opts.workingDir)
	if err != nil {
		return fmt.Errorf("run task: %w", err)
	}

	if opts.timeout != nil && *opts.timeout != 0 {
		ctx, cancel = context.WithTimeoutCause(ctx, *opts.timeout, ErrTimeout)
		defer cancel()
	}

	frame := newTopFrame(taskName, opts.module, opts.env, opts.args)

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

	err = runTask(ctx, taskName, tasks, frame, termChannel, termWaitGroup)
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

func newTopFrame(taskName string, module string, env []string, args []string) *Frame {
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

	frame.SetVars(task.GetEnv())

	if task.GetInheritEnv() {
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

		for _, cmd := range task.GetCmds() {
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
) error {
	var cmdErr error

	logger := log.FromContext(ctx).With().Str("cur", taskName).Logger()
	ctx = logger.WithContext(ctx)

	logger.Info().Msg("task")

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

	expandTaskProperties(task, frame.ExpandMapping())

	for cmdIdx, cmd := range task.GetCmds() {
		switch {
		case cmd.Task != nil:
			cmdErr = runTask(ctx, *cmd.Task, tasks, frame, termChannel, termWaitGroup)
		case cmd.EmbeddedShell:
			logger.Info().Str("run", "shell").Msg(displayCommand(cmd.Cmd))

			scriptName := fmt.Sprintf("%s[%d]", taskName, cmdIdx)
			cmdErr = runShell(ctx, scriptName, cmd.Cmd, task.GetWorkingDir(), frame.EnvList())
		default:
			logger.Info().Str("run", "command").Msg(displayCommand(cmd.Cmd))

			cmdErr = runCmd(ctx, cmd.Cmd, task.GetWorkingDir(), frame.EnvList())
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
func runCmd(ctx context.Context, command []string, dir string, environ []string) error {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = environ

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ee := (&exec.ExitError{})

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

// runShell runs an arbitrary shell command or script with an mvdan.cc shell interpreter, so called
// "embedded shell" in tpkl. cf. https://github.com/mvdan/sh
func runShell(ctx context.Context, taskName string, command []string, dir string, environ []string) error {
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

func displayCommand(command []string) string {
	var displayCommand strings.Builder

	const displayLen = 72

	NLRegexp := regexp.MustCompile(`\n+`)

	displayCommand.Grow(displayLen)

	for _, cmd := range command {
		c := NLRegexp.ReplaceAllString(cmd, " ")

		displayCommand.WriteByte(' ')
		displayCommand.WriteString(c)
	}

	return truncate(strings.TrimSpace(displayCommand.String()), displayLen)
}

func truncate(text string, maxLen int) string {
	lastSpace := maxLen
	length := 0

	for i, r := range text {
		if unicode.IsSpace(r) {
			lastSpace = i
		}

		length++

		if length > maxLen {
			return text[:lastSpace] + "..."
		}
	}

	return text
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
	files := task.GetFiles()

	if len(files) == 0 {
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

	for key, file := range files {
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

func expandTaskProperties(task tpkl.Task, mapping func(string) string) {
	taskI, _ := task.(tpkl.TaskImpl)

	for _, cmd := range taskI.Cmds {
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
