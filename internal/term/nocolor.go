package term

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

// NoColor returns true if *no* color output should be done on the
// terminal, if any, associated wth the io.Writer passed as first
// argument.
// It uses github.com/charmbracelet/colorprofile.
//
// https://github.com/charmbracelet/colorprofile
// https://bixense.com/clicolors/
// https://no-color.org/
func NoColor(output io.Writer) bool {
	p := colorprofile.Detect(output, os.Environ())
	switch p { //nolint:exhaustive
	case colorprofile.NoTTY, colorprofile.Ascii:
		return true
	default:
		return false
	}
}
