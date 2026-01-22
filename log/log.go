// Package log implements logging for pkl
package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mgutz/ansi"
	"github.com/rs/zerolog"
	"golang.org/x/term"
)

// FromContext retrieve a logger from a context.
func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// New returns a logger for a command.
func New(cmd string, verbosity int) zerolog.Logger {
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
			writer.FieldsExclude = []string{"tpkl"}
			writer.FieldsOrder = []string{"task", "cur"}
			writer.NoColor = noColor
		},
	)
	logger := zerolog.New(loggerOut).Level(level).With().Timestamp().Str("tpkl", cmd).Logger()

	return logger
}

// AsFatal generate a Fatal-like log with a message.
func AsFatal(logger *zerolog.Logger, msg string) {
	logger.WithLevel(zerolog.FatalLevel).Msg(msg)
}

func logPartsOrder() []string {
	return []string{
		zerolog.TimestampFieldName,
		zerolog.LevelFieldName,
		zerolog.CallerFieldName,
		"tpkl",
		// "task",
		zerolog.MessageFieldName,
	}
}

func getFormatPartValueByName(noColor bool) func(i any, s string) string {
	var cyan func(string) string
	if !noColor {
		cyan = ansi.ColorFunc("cyan")
	}

	return func(value any, part string) string {
		var ret, key string

		if part == "tpkl" {
			if !noColor {
				key = cyan(part + "=")
			} else {
				key = part + "="
			}

			ret = fmt.Sprintf("%s%s", key, value)
		}

		return ret
	}
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

	return !term.IsTerminal(int(os.Stdout.Fd()))
}
