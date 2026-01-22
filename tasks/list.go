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
	return enumarg.New([]string{"name", "json"}, "name")
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
func List(ctx context.Context, writer io.Writer, format string, options ...ListOption) error {
	var err error

	opts := &listOptions{}
	for _, opt := range options {
		opt.setListOption(opts)
	}

	switch format {
	case "name":
	case "json":
	default:
		return fmt.Errorf("list tasks: %w: `%s`", ErrUnknownOption, format)
	}

	opts.module, err = useModule(ctx, opts.module, "") // XXX support working-dir
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	frame := newTopFrame("", opts.module, opts.env, nil)

	tasks, err := ModuleTasks(ctx, opts.module, WithPklEnv(frame.EnvList()), WithPklProperties(opts.properties),
		WithPklProperties([]string{identifierPrefix + "LIST_COMMAND_RUNNING"}))
	if err != nil {
		return err
	}

	switch format {
	case "name":
		return listName(tasks, writer)
	case "json":
		return listJSON(tasks, writer)
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
