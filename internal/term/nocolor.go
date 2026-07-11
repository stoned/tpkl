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
// https://force-color.org/
// https://no-color.org/
// https://web.archive.org/web/20260616201813/https://no-color.org/
// https://bixense.com/clicolors/
func NoColor(output io.Writer) bool {
	p := colorprofile.Detect(output, os.Environ())
	switch p { //nolint:exhaustive
	case colorprofile.NoTTY, colorprofile.Ascii:
		return true
	default:
		return false
	}
}
