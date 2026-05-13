package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/stoned/tpkl/internal/enumarg"
)

// FormatEnumArg returns github.com/spf13/pflag to record
// a CLI flag to set List()'s format.
func FormatEnumArg() *enumarg.EnumArg {
	return enumarg.New([]string{"name", "json", "summary"}, "name") //nolint:goconst
}

// Options for List().
type listOptions struct {
	env        []string
	module     string
	properties []string
}

// ListOption is  List()'s options interface.
type ListOption interface {
	setListOption(o *listOptions)
}

// Set env List()'s option.
func (e *envOption) setListOption(o *listOptions) {
	o.env = e.env
}

// Set module List()'s option.
func (m *moduleOption) setListOption(o *listOptions) {
	o.module = m.module
}

// Set properties List()'s option.
func (p *propertiesOption) setListOption(o *listOptions) {
	o.properties = p.properties
}

// List lists tasks defined in a Pkl module.
func List(ctx context.Context, writer io.Writer, format string, options ...ListOption) error { //nolint:cyclop
	var err error

	opts := &listOptions{}
	for _, opt := range options {
		opt.setListOption(opts)
	}

	switch format {
	case "json":
	case "name":
	case "summary":
	default:
		return fmt.Errorf("list tasks: %w: `%s`", ErrUnknownOption, format)
	}

	opts.module, err = UseModule(ctx, opts.module, "") // XXX support working-dir
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	frame := NewTopFrame("", opts.module, opts.env, nil)

	tasks, err := ModuleTasks(ctx, opts.module, WithPklEnv(frame.EnvList()), WithPklProperties(opts.properties),
		WithPropertyListCommandRunning())
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return listJSON(tasks, writer)
	case "name":
		return listName(tasks, writer)
	case "summary":
		return list(tasks, writer)
	}

	return nil
}

func listJSON(tasks Tasks, writer io.Writer) error {
	b, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrJSONMarshal, err)
	}

	_, err = fmt.Fprintln(writer, string(b))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	return nil
}

func listName(tasks Tasks, writer io.Writer) error {
	names := slices.Sorted(maps.Keys(tasks))

	for _, t := range names {
		_, err := fmt.Fprintln(writer, t)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIO, err)
		}
	}

	return nil
}

func list(tasks Tasks, writer io.Writer) error {
	names := slices.Sorted(maps.Keys(tasks))

	nameMaxLen := 0

	for _, name := range names {
		l := len(name)
		if l > nameMaxLen {
			nameMaxLen = l
		}
	}

	for _, name := range names {
		var line string
		if tasks[name].Summary != "" {
			line = fmt.Sprintf("%-*s  %s", nameMaxLen, name, tasks[name].Summary)
		} else {
			line = name
		}

		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIO, err)
		}
	}

	return nil
}
