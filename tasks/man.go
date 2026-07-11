package tasks

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
	"github.com/gerow/pager"
	"github.com/stoned/tpkl/internal/markdown"
	"github.com/stoned/tpkl/log"
)

// ManOptions are options for Man().
type manOptions struct {
	evalOptions

	nopager bool
}

// ManOption is Man()'s options interface.
type ManOption interface {
	setManOption(opts *manOptions)
}

func (o *envOption) setManOption(opts *manOptions) {
	opts.env = o.env
}

func (o *moduleOption) setManOption(opts *manOptions) {
	opts.module = o.module
}

func (o *nopagerOption) setManOption(opts *manOptions) {
	opts.nopager = o.nopager
}

func (o *propertiesOption) setManOption(opts *manOptions) {
	opts.properties = o.properties
}

// Man show a task's documentation.
func Man(ctx context.Context, writer io.Writer, taskName string, options ...ManOption) error { //nolint:cyclop,funlen
	var err error

	opts := &manOptions{}
	for _, opt := range options {
		opt.setManOption(opts)
	}

	logger := log.FromContext(ctx).With().Str("task", taskName).Logger()
	ctx = logger.WithContext(ctx)

	opts.module, err = UseModule(ctx, opts.module, opts.workingDir)
	if err != nil {
		return fmt.Errorf("man task: %w", err)
	}

	frame := NewTopFrame(taskName, opts.module, opts.env, []string{})

	tasks, err := ModuleTasks(ctx, opts.module, WithPklEnv(frame.EnvList()),
		WithPklProperties(opts.properties))
	if err != nil {
		return err
	}

	if _, ok := tasks[taskName]; !ok {
		return fmt.Errorf("%w: `%s`", ErrUnknownTask, taskName)
	}

	if tasks[taskName].Doc == nil || *tasks[taskName].Doc == "" {
		fmt.Printf("No documentation for %s\n", taskName)

		return nil
	}

	rawDoc := *tasks[taskName].Doc

	renderer, err := markdown.NewRenderer(writer, 0)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	doc, err := renderer.Render(rawDoc)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	// create a colorprofiled writer **before** the given writer is
	// eventually piped into a pager
	out := colorprofile.NewWriter(writer, os.Environ())

	if !opts.nopager {
		err = pager.Open()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIO, err)
		}

		defer func() {
			err := pager.Close()
			if err != nil {
				fmt.Printf("error closing pager: %s: %s\n", ErrIO, err)
			}
		}()
	}

	_, err = fmt.Fprintln(out, doc)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	return nil
}
