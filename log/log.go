// Package log implements logging for pkl
package log

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-logfmt/logfmt"
	"github.com/mgutz/ansi"
	"github.com/rs/zerolog"
	"golang.org/x/term"
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
			writer.TimeFormat = "15:04:05.000"
			writer.NoColor = noColor
			writer.Out = os.Stderr
			writer.FormatPartValueByName = getFormatPartValueByName(noColor)
			writer.FieldsExclude = []string{"_header", "call", "cmd", "cur", "shell", "task", "tpkl"}
			writer.PartsOrder = []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				"tpkl",
				"task",
				"cur",
				"call",
				"cmd",
				"shell",
				zerolog.MessageFieldName,
				"_header",
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

// Cmd logs command to be executed.
func Cmd(ctx context.Context, command []string) {
	logger := FromContext(ctx)
	logLevel := logger.GetLevel()

	switch {
	case logLevel > zerolog.InfoLevel:
		return

	case logLevel == zerolog.InfoLevel:
		logger.Info().Str("cmd", truncatedFoldedStrings(command)).Send()

	case logLevel == zerolog.DebugLevel:
		logger.Debug().Str("cmd", foldedStrings(command)).Send()

	case logLevel <= zerolog.TraceLevel:
		// ┌ U+250C BOX DRAWINGS LIGHT DOWN AND RIGHT
		// ╵ U+2575 BOX DRAWINGS LIGHT UP
		logger.Trace().Str("_header", "cmd").Msg("\u250c")
		logger.Trace().Msg("\u2575 " + strings.Join(command, " "))
	}
}

// ShellCmd logs a command to be executed by the embedded shell.
func ShellCmd(ctx context.Context, command []string) {
	logger := FromContext(ctx)
	logLevel := logger.GetLevel()

	switch {
	case logLevel > zerolog.InfoLevel:
		return

	case logLevel == zerolog.InfoLevel:
		logger.Info().Str("shell", truncatedFoldedStrings(command)).Send()

	case logLevel == zerolog.DebugLevel:
		logger.Debug().Str("shell", foldedStrings(command)).Send()

	case logLevel <= zerolog.TraceLevel:
		lines := strings.Split(command[0], "\n")

		for idx, line := range lines {
			// ┌ U+250C BOX DRAWINGS LIGHT DOWN AND RIGHT
			// │ U+2502 BOX DRAWINGS LIGHT VERTICAL
			// ╵ U+2575 BOX DRAWINGS LIGHT UP
			if idx == 0 {
				logger.Trace().Strs("args", command[1:]).Send()
				logger.Trace().Str("_header", "shell").Msg("\u250c")
			}

			if idx == len(lines)-1 {
				logger.Trace().Msg("\u2575 " + line)
			} else {
				logger.Trace().Msg("\u2502 " + line)
			}
		}
	}
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

func getFormatPartValueByName(noColor bool) func(i any, s string) string {
	var bold, cyan func(string) string
	if !noColor {
		bold = ansi.ColorFunc("white+b")
		cyan = ansi.ColorFunc("cyan")
	}

	return func(value any, part string) string {
		var ret, key string

		if value == nil {
			return ""
		}

		switch part {
		case "_header":
			valueString := fmt.Sprintf("%s", value)
			if !noColor {
				ret = bold(valueString)
			} else {
				ret = valueString
			}

		case "cur", "task", "tpkl":
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
