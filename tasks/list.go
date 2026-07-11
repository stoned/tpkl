package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/stoned/tpkl/internal/enumarg"
	"github.com/stoned/tpkl/internal/markdown"
)

// FormatEnumArg returns github.com/spf13/pflag to record
// a CLI flag to set List()'s format.
func FormatEnumArg() *enumarg.EnumArg {
	return enumarg.New([]string{"name", "json", "summary", "summary-raw"}, "name") //nolint:goconst
}

// Options for List().
type listOptions struct {
	evalOptions
}

// ListOption is  List()'s options interface.
type ListOption interface {
	setListOption(opts *listOptions)
}

// Set env List()'s option.
func (o *envOption) setListOption(opts *listOptions) {
	opts.env = o.env
}

// Set module List()'s option.
func (o *moduleOption) setListOption(opts *listOptions) {
	opts.module = o.module
}

// Set properties List()'s option.
func (o *propertiesOption) setListOption(opts *listOptions) {
	opts.properties = o.properties
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
	case "summary-raw":
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
		return listWithSummary(tasks, writer)
	case "summary-raw":
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
	if len(tasks) == 0 {
		return nil
	}

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

func listWithSummary(tasks Tasks, writer io.Writer) error {
	if len(tasks) == 0 {
		return nil
	}

	names := slices.Sorted(maps.Keys(tasks))

	nameMaxLen := 0

	for _, name := range names {
		l := len(name)
		if l > nameMaxLen {
			nameMaxLen = l
		}
	}

	renderer, err := markdown.NewRenderer(writer, nameMaxLen+2) //nolint:mnd
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	out := table.New().
		BorderBottom(false).
		BorderColumn(false).
		BorderLeft(false).
		BorderRight(false).
		BorderRow(false).
		BorderTop(false).
		StyleFunc(func(_, _ int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 0)
		})

	for _, name := range names {
		if tasks[name].Summary != "" {
			info, err := renderer.Render(tasks[name].Summary)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrIO, err)
			}

			out.Row(name, strings.Trim(info, "\n"))
		} else {
			out.Row(name)
		}
	}

	_, err = lipgloss.Fprintln(writer, out)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	return nil
}
