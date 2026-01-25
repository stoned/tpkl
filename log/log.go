// Package log implements logging for pkl
package log

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-logfmt/logfmt"
	"github.com/mgutz/ansi"
	"github.com/rs/zerolog"
	"golang.org/x/term"
)

type (
	contextKey byte
	logBuilder func() zerolog.Logger
)

const (
	traceVerbosity = 3
)

const (
	builderContextKey contextKey = iota
)

// FromContext retrieve a logger from a context.
func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// ContextWithLogger creates a logger builder, a logger, store them in
// a context and returns the context and the logger.
func ContextWithLogger(ctx context.Context, cmd string, verbosity int) (context.Context, *zerolog.Logger) {
	logBuilder := builder(cmd, verbosity)
	logger := logBuilder()

	ctx = logger.WithContext(ctx)
	ctx = contextWithBuilder(ctx, logBuilder)

	return ctx, &logger
}

func contextWithBuilder(ctx context.Context, builder logBuilder) context.Context {
	return context.WithValue(ctx, builderContextKey, builder)
}

func builderFromContext(ctx context.Context) logBuilder {
	if val := ctx.Value(builderContextKey); val != nil {
		if builder, ok := val.(logBuilder); ok {
			return builder
		}
	}

	// in case no builder was found in the context return a new one with a
	// *high* verbosity level
	return builder("no-command", traceVerbosity)
}

func builder(cmd string, verbosity int) func() zerolog.Logger {
	return func() zerolog.Logger {
		var (
			level   zerolog.Level
			noColor bool
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

		noColor = EnvNoColor()

		// https://github.com/rs/zerolog/issues/114
		zerolog.TimeFieldFormat = time.RFC3339Nano

		loggerOut := zerolog.NewConsoleWriter(
			func(writer *zerolog.ConsoleWriter) {
				writer.PartsOrder = logPartsOrder()
				writer.TimeFormat = "15:04:05.000"
				writer.FormatPartValueByName = getFormatPartValueByName(noColor)
				writer.FieldsExclude = []string{"call", "cmd", "shell", "task", "tpkl"}
				writer.FieldsOrder = []string{"cur"}
				writer.NoColor = noColor
				writer.Out = os.Stderr
			},
		)
		logger := zerolog.New(loggerOut).Level(level).With().Timestamp().Str("tpkl", cmd).Logger()

		return logger
	}
}

// AsFatal generate a Fatal-like log with a message.
func AsFatal(logger *zerolog.Logger, msg string) {
	logger.WithLevel(zerolog.FatalLevel).Msg(msg)
}

// DebugCmd log at debug level a command to be executed.
func DebugCmd(ctx context.Context, command []string) {
	logBuilder := builderFromContext(ctx)
	logger := logBuilder()

	if logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	// │ U+2502 BOX DRAWINGS LIGHT VERTICAL
	logger.Debug().Msg("\u2502 " + strings.Join(command, " "))
}

// DebugShell log at debug level a command to be executed by the
// embedded shell.
func DebugShell(ctx context.Context, command []string) {
	logBuilder := builderFromContext(ctx)
	logger := logBuilder()

	if logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	logger.Debug().Strs("args", command[1:]).Send()

	lines := strings.Split(command[0], "\n")

	if len(lines) == 1 {
		logger.Debug().Msg("\u2502 " + command[0])
	} else {
		for idx, line := range lines {
			// ╷ U+2577 BOX DRAWINGS LIGHT DOWN
			// │ U+2502 BOX DRAWINGS LIGHT VERTICAL
			// ╵ U+2575 BOX DRAWINGS LIGHT UP
			switch {
			case idx == 0:
				logger.Debug().Msg("\u2577 " + line)
			case idx == len(lines)-1:
				logger.Debug().Msg("\u2575 " + line)
			default:
				logger.Debug().Msg("\u2502 " + line)
			}
		}
	}
}

func logPartsOrder() []string {
	return []string{
		zerolog.TimestampFieldName,
		zerolog.LevelFieldName,
		zerolog.CallerFieldName,
		"tpkl",
		"task",
		"call",
		"cmd",
		"shell",
		zerolog.MessageFieldName,
	}
}

func getFormatPartValueByName(noColor bool) func(i any, s string) string {
	var bold, cyan func(string) string
	if !noColor {
		bold = ansi.ColorFunc("white+b")
		cyan = ansi.ColorFunc("cyan")
	}

	return func(value any, part string) string {
		var ret, key string

		encodeKeyval := func(key, val any) string {
			buff := new(strings.Builder)
			encoder := logfmt.NewEncoder(buff)

			err := encoder.EncodeKeyval(key, val)
			if err != nil {
				_ = encoder.EncodeKeyval(key, fmt.Sprintf("%+v", val))
			}

			return buff.String()
		}

		switch part {
		case "task", "tpkl":
			if value == nil {
				return ""
			}

			if !noColor {
				key = cyan(part + "=")
			} else {
				key = part + "="
			}

			valueString := fmt.Sprintf("%s", value)
			if quotep(valueString) {
				valueString = strconv.Quote(valueString)
			}

			ret = key + valueString

		case "call", "cmd", "shell":
			if value == nil {
				return ""
			}

			if !noColor {
				ret = bold(encodeKeyval(part, value))
			} else {
				ret = encodeKeyval(part, value)
			}
		}

		return ret
	}
}

func quotep(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) || !strconv.IsPrint(r) || strings.ContainsAny(string(r), "\\\"") {
			return true
		}
	}

	return false
}

// EnvNoColor lookups relevant environment variables and if stdout
// is a terminal and returns true if *no* color output should be done.
//
// https://bixense.com/clicolors/
// https://no-color.org/
func EnvNoColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}

	force, ok := os.LookupEnv("CLICOLOR_FORCE")
	if ok && force != "0" {
		return false
	}

	if (ok && force == "0") || os.Getenv("CLICOLOR") == "0" {
		return true
	}

	return !term.IsTerminal(int(os.Stderr.Fd()))
}
