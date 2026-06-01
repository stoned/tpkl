// Package log implements logging for pkl
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/go-logfmt/logfmt"
	"github.com/rs/zerolog"
)

const (
	lineIndent = " "
)

// FromContext retrieve a logger from a context.
func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// ContextWithLogger creates a logger, stores it in a context and
// returns the context and the logger.
func ContextWithLogger(ctx context.Context, cmd string, verbosity int) (context.Context, *zerolog.Logger) {
	logger := Builder(cmd, verbosity)

	ctx = logger.WithContext(ctx)

	return ctx, &logger
}

// Builder makes a new zerolog Logger.
func Builder(cmd string, verbosity int) zerolog.Logger {
	var (
		level   zerolog.Level
		nocolor bool
	)

	switch {
	case verbosity > 2: //nolint:mnd
		level = zerolog.TraceLevel
	case verbosity > 1:
		level = zerolog.DebugLevel
	case verbosity > 0:
		level = zerolog.InfoLevel
	default:
		level = zerolog.WarnLevel
	}

	nocolor = noColor(os.Stderr)

	// https://github.com/rs/zerolog/issues/114
	zerolog.TimeFieldFormat = time.RFC3339Nano

	loggerOut := zerolog.NewConsoleWriter(
		func(writer *zerolog.ConsoleWriter) {
			writer.TimeFormat = "15:04:05.000"
			writer.NoColor = nocolor
			writer.Out = os.Stderr
			writer.FormatPartValueByName = getFormatPartValueByName(nocolor)
			writer.FieldsExclude = []string{"call", "cmd", "cur", "shell", "task", "tpkl"}
			writer.PartsOrder = []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				"tpkl",
				"task",
				"cur",
				"call",
				"cmd",
				"shell",
				zerolog.MessageFieldName,
			}
		},
	)
	logger := zerolog.New(loggerOut).Level(level).With().Timestamp().Str("tpkl", cmd).Logger()

	return logger
}

// AsFatal generate a Fatal-like log with a message.
func AsFatal(logger *zerolog.Logger, msg string) {
	logger.WithLevel(zerolog.FatalLevel).Msg(msg)
}

// Cmd logs a command to be executed.
func Cmd(ctx context.Context, command []string) {
	var cmd string

	logger := FromContext(ctx)
	logLevel := logger.GetLevel()

	switch {
	case logLevel > zerolog.InfoLevel:
		return

	case logLevel == zerolog.InfoLevel:
		cmd = truncatedFoldedStrings(command)

	case logLevel == zerolog.DebugLevel:
		cmd = foldedStrings(command)

	case logLevel <= zerolog.TraceLevel:
		cmd = strings.Join(command, " ")
	}

	logger.Info().Str("cmd", cmd).Send()
}

// Shell logs a command to be executed by the embedded shell.
func Shell(ctx context.Context, command []string) {
	var fields []any

	logger := FromContext(ctx)
	logLevel := logger.GetLevel()

	switch {
	case logLevel > zerolog.InfoLevel:
		return

	case logLevel == zerolog.InfoLevel:
		fields = append(fields, "shell", truncatedFoldedStrings(command[0:1]))

	case logLevel == zerolog.DebugLevel:
		fields = append(fields, "shell", foldedStrings(command[0:1]))

	case logLevel <= zerolog.TraceLevel:
		fields = append(fields, "shell", command[0], "args", command[1:])
	}

	logger.Info().Fields(fields).Send()
}

func encodeKeyval(key, val any) string {
	buff := new(strings.Builder)
	encoder := logfmt.NewEncoder(buff)

	err := encoder.EncodeKeyval(key, val)
	if err != nil {
		_ = encoder.EncodeKeyval(key, fmt.Sprintf("%+v", val))
	}

	return buff.String()
}

func getFormatPartValueByName(nocolor bool) func(i any, s string) string {
	var cyan, yellow, faint func(string) string

	if !nocolor {
		yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Yellow)

		yellow = func(s string) string {
			return yellowStyle.Render(s)
		}

		cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Cyan)

		cyan = func(s string) string {
			return cyanStyle.Render(s)
		}

		faintStyle := lipgloss.NewStyle().Faint(true)

		faint = func(s string) string {
			return faintStyle.Render(s)
		}
	} else {
		yellow = func(s string) string { return s }
		cyan = func(s string) string { return s }
		faint = func(s string) string { return s }
	}

	return func(value any, part string) string {
		var ret, key string

		if value == nil {
			return ""
		}

		switch part {
		case "cur", "task", "tpkl":
			key = cyan(part + "=")

			val := fmt.Sprintf("%s", value)
			if quotep(val) {
				val = strconv.Quote(val)
			}

			ret = key + val

		case "call", "cmd":
			ret = yellow(encodeKeyval(part, value))

		case "shell":
			val := fmt.Sprintf("%s", value)
			if strings.Contains(val, "\n") {
				ret = formatMultiline(val, part, faint, yellow)
			} else {
				ret = yellow(encodeKeyval(part, value))
			}
		}

		return ret
	}
}

func formatMultiline(value string, part string, faint, colorize func(string) string) string {
	var buf strings.Builder

	linePrefix := faint(lineIndent + "\u2502 ")

	buf.WriteByte('\n')
	buf.WriteString(lineIndent)
	buf.WriteString(colorize(part + "="))
	buf.WriteByte('\n')

	for line := range strings.SplitSeq(value, "\n") {
		buf.WriteString(linePrefix)
		buf.WriteString(colorize(line))
		buf.WriteByte('\n')
	}

	return buf.String()
}

func quotep(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) || !strconv.IsPrint(r) || strings.ContainsAny(string(r), "\\\"") {
			return true
		}
	}

	return false
}

// noColor, using github.com/charmbracelet/colorprofile, lookups
// relevant environment variables and terminal properties and returns
// true if *no* color output should be done.
//
// https://github.com/charmbracelet/colorprofile
// https://bixense.com/clicolors/
// https://no-color.org/
func noColor(output io.Writer) bool {
	p := colorprofile.Detect(output, os.Environ())
	switch p { //nolint:exhaustive
	case colorprofile.NoTTY, colorprofile.Ascii:
		return true
	default:
		return false
	}
}

func foldedStrings(chunks []string) string {
	var result strings.Builder

	NLRegexp := regexp.MustCompile(`\n+`)

	for _, chunk := range chunks {
		folded := NLRegexp.ReplaceAllString(chunk, " ")

		result.WriteByte(' ')
		result.WriteString(folded)
	}

	return strings.TrimSpace(result.String())
}

func truncatedFoldedStrings(chunks []string) string {
	const maxLen = 72

	return truncate(foldedStrings(chunks), maxLen)
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
