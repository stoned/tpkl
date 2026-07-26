// Package tasks handles tasks
package tasks

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/apple/pkl-go/pkl"
	"github.com/stoned/tpkl/extreaders"
	"github.com/stoned/tpkl/internal/spath"
	"github.com/stoned/tpkl/log"
	"github.com/stoned/tpkl/modules/tpkl"
)

// tasksExpression is a Pkl expression returning `tasks` from the evaluated module.
const tasksExpression = `tasks`

// identifierPrefix is the prefix for environment variables and properties set by tpkl.
const identifierPrefix = "TPKL_"

// filesCountVarNameSuffix is the suffix for the environment variable holding
// the number of files defined by the currently executing Task.
const filesCountVarNameSuffix = "FILES_COUNT"

// moduleFilename is the default Pkl module filename in which tasks
// are searched for.
const moduleFilename = "tasks.pkl"

var (
	// ErrEvaluateExpr signals an error while evaluating the Pkl module.
	ErrEvaluateExpr = errors.New("error evaluating expression in module")
	// ErrIO signals an I/O error.
	ErrIO = errors.New("I/O error")
	// ErrJSONMarshal signals an error marshaling as JSON.
	ErrJSONMarshal = errors.New("error marshaling as JSON")
	// ErrNewEvaluator signals an error creating a Pkl evaluator.
	ErrNewEvaluator = errors.New("error creating Pkl evaluator")
	// ErrNoModule signals an error loading/searching for a Pkl module.
	ErrNoModule = errors.New("no module")
	// ErrNoProject signals an error searching for a PklProject file.
	ErrNoProject = errors.New("error searching for PklProject")
	// ErrTaskCycle signals a call cycle between tasks.
	ErrTaskCycle = errors.New("tasks cycle")
	// ErrTaskFile signals an error with a task file.
	ErrTaskFile = errors.New("error with task file")
	// ErrTimeout signals a timed out task.
	ErrTimeout = errors.New("task timed out")
	// ErrUnknownTask signals the requested tasks was not found in the Pkl module.
	ErrUnknownTask = errors.New("unknown task")
	// ErrUnknownOption signals an unknown option.
	ErrUnknownOption = errors.New("unknown option")
)

// EvalError signals a Pkl evaluation error.
type EvalError struct {
	module string
	expr   string
	err    error
}

func (r *EvalError) Error() string {
	var sep string
	if r.err != nil && strings.HasPrefix(r.err.Error(), "–– Pkl Error ––") {
		sep = "\n"
	} else {
		sep = " "
	}

	return fmt.Sprintf("error evaluating expression `%s` in module `%s`:%s%s", r.expr, r.module, sep, r.err)
}

// Is implements the interface expected by errors.Is for EvalError.
func (r *EvalError) Is(err error) bool {
	if err == nil {
		return false
	}

	var evalErr *EvalError

	ok := errors.As(err, &evalErr)

	return ok
}

// Unwrap implements the interface expected by errors.Unwrap for EvalError.
func (r *EvalError) Unwrap() error {
	return r.err
}

// Tasks maps task names to their properties as tpkl.Task values.
type Tasks map[string]tpkl.Task

// CmdError wraps an error encountered running a command.
type CmdError struct {
	ExitCode int
	Err      error
}

// NewCmdError creates a CmdError error.
func NewCmdError(exitCode int, err error) *CmdError {
	return &CmdError{
		ExitCode: exitCode,
		Err:      err,
	}
}

func (e *CmdError) Error() string {
	errStr := e.Err.Error()

	if strings.HasPrefix(errStr, "exit status ") {
		return errStr
	}

	return fmt.Sprintf("exit status %d: %s", e.ExitCode, errStr)
}

// WithPklEnv returns a pkl.Evaluator options function to set environment variables.
func WithPklEnv(vars []string) func(*pkl.EvaluatorOptions) {
	return func(opts *pkl.EvaluatorOptions) {
		if len(vars) == 0 {
			return
		}

		if opts.Env == nil {
			opts.Env = make(map[string]string)
		}

		for _, variable := range vars {
			name, val, found := strings.Cut(variable, "=")
			if !found {
				val = ""
			}

			opts.Env[name] = val
		}
	}
}

// WithPklProperties returns a pkl.Evaluator options function to set properties.
func WithPklProperties(properties []string) func(*pkl.EvaluatorOptions) {
	return func(opts *pkl.EvaluatorOptions) {
		if len(properties) == 0 {
			return
		}

		if opts.Properties == nil {
			opts.Properties = make(map[string]string)
		}

		for _, prop := range properties {
			name, val, found := strings.Cut(prop, "=")
			if !found {
				val = "true"
			}

			opts.Properties[name] = val
		}
	}
}

// WithPropertyListCommandRunning returns a pkl.Evaluator options function to set property "LIST_COMMAND_RUNNING".
func WithPropertyListCommandRunning() func(*pkl.EvaluatorOptions) {
	name := identifierPrefix + "LIST_COMMAND_RUNNING"

	return func(opts *pkl.EvaluatorOptions) {
		if opts.Properties == nil {
			opts.Properties = make(map[string]string)
		}

		opts.Properties[name] = "true"
	}
}

// evalOptions are options for commands implementations which run a Pkl evaluation.
type evalOptions struct {
	envOption
	moduleOption
	propertiesOption
	workingDirOption
}

// WithArgs initializes a struct to define an "arguments option".
func WithArgs(args []string) *argsOption {
	return &argsOption{args}
}

type argsOption struct {
	args []string
}

// WithDryrun initializes a struct to define a "dryrun option".
func WithDryrun(dryrun bool) *dryrunOption {
	return &dryrunOption{dryrun}
}

type dryrunOption struct {
	dryrun bool
}

// WithEnv initializes a struct to define an "env option".
func WithEnv(vars []string) *envOption {
	return &envOption{vars}
}

type envOption struct {
	env []string
}

// WithModule initializes a struct to define a "module option".
func WithModule(module string) *moduleOption {
	return &moduleOption{module}
}

type moduleOption struct {
	module string
}

// WithNoPager initializes a struct to define a "nopager option".
func WithNoPager(nopager bool) *nopagerOption {
	return &nopagerOption{nopager}
}

type nopagerOption struct {
	nopager bool
}

// WithProperties initializes a struct to define a "properties option".
func WithProperties(properties []string) *propertiesOption {
	return &propertiesOption{properties}
}

type propertiesOption struct {
	properties []string
}

// WithTimeout initializes a struct to define a "timeout option".
func WithTimeout(timeout time.Duration) *timeoutOption {
	return &timeoutOption{timeout}
}

type timeoutOption struct {
	timeout time.Duration
}

// WithWorkingDir initializes a struct to define a "workingDir option".
func WithWorkingDir(path string) *workingDirOption {
	return &workingDirOption{path}
}

type workingDirOption struct {
	workingDir string
}

// ModuleTasks returns tpkl Tasks defined in module.
func ModuleTasks(ctx context.Context, module string,
	options ...func(*pkl.EvaluatorOptions),
) (Tasks, error) {
	var (
		err       error
		evaluator pkl.Evaluator
	)

	out := make(Tasks)

	manager := pkl.NewEvaluatorManager()

	defer (func(mgr pkl.EvaluatorManager) {
		_ = mgr.Close()
	})(manager)

	withValsResourcesReader, err := extreaders.NewWithValsResourceReader()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewEvaluator, err)
	}

	opts := make([]func(*pkl.EvaluatorOptions), 0, 3+len(options)) //nolint:mnd

	opts = append(opts,
		pkl.PreconfiguredOptions,
		pkl.WithModuleReader(extreaders.ModuleReader{}),
		withValsResourcesReader,
	)
	opts = append(opts, options...)

	// XXX add support for --project-dir= command line option
	projectDir, err := findProjectDir(moduleDir(module))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewEvaluator, err)
	}

	if projectDir == nil {
		evaluator, err = manager.NewEvaluator(ctx, opts...)
	} else {
		logger := log.FromContext(ctx).With().Logger()
		logger.Debug().Str("projectdir", projectDir.String()).Send()
		evaluator, err = manager.NewProjectEvaluator(ctx, projectDir, opts...)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewEvaluator, err)
	}

	err = evaluator.EvaluateExpression(ctx,
		pkl.FileSource(module), tasksExpression, &out)
	if err != nil {
		e := &EvalError{module, tasksExpression, err}
		if context.Cause(ctx) != nil {
			return nil, fmt.Errorf("%w: %w", context.Cause(ctx), e)
		}

		return nil, e
	}

	return out, nil
}

// SetWorkingDir change the current working directory.
func SetWorkingDir(ctx context.Context, wd string) (string, error) {
	workingDir, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIO, err)
	}

	err = os.Chdir(workingDir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIO, err)
	}

	logger := log.FromContext(ctx)
	logger.Debug().Str("workdir", workingDir).Send()

	return workingDir, nil
}

// UseModule searches for the tasks PKL module to use and advertise it if requested.
func UseModule(ctx context.Context, module string, workingDir string) (string, error) {
	var err error

	if module == "" {
		module, err = findModule(workingDir)
		if err != nil {
			return "", err
		}
	} else {
		if !filepath.IsAbs(module) {
			module = filepath.Join(workingDir, module)
		}

		module, err = filepath.Abs(module)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrIO, err)
		}
	}

	logger := log.FromContext(ctx)
	logger.Debug().Str("module", module).Send()

	return module, nil
}

// findModule returns the tpkl Tasks module path, walking the filesystem
// up to the root of the filesystem, either from the given directory, or
// else from the current working directory.
func findModule(fromDir string) (string, error) {
	var (
		dir, module string
		err         error
	)

	dir, err = filepath.Abs(fromDir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoModule, err)
	}

	module, err = spath.FindUp(moduleFilename, dir, os.DirFS("/"))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoModule, err)
	}

	return module, nil
}

// findProjectDir returns the first directory containing a PklProject
// file, walking the filesystem up to the root of the filesystem, either
// from the given directory, or else from the current working directory.
func findProjectDir(fromDir string) (*url.URL, error) {
	var (
		dir, projectDir, projectFile string
		err                          error
	)

	if fromDir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNoProject, err)
		}
	} else {
		dir, err = filepath.Abs(fromDir)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNoProject, err)
		}
	}

	projectFile, err = spath.FindUp("PklProject", dir, os.DirFS("/"))
	if err != nil {
		if errors.Is(err, spath.ErrFindUp) {
			return nil, nil //nolint:nilnil
		}

		return nil, fmt.Errorf("%w: %w", ErrNoProject, err)
	}

	projectDir = filepath.Dir(projectFile)

	return &url.URL{Scheme: "file", Path: projectDir}, nil
}

func moduleDir(module string) string {
	_, err := url.Parse(module)
	if err == nil {
		return path.Dir(module)
	}

	return filepath.Dir(module)
}
