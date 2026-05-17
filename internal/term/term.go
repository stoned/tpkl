// Package term provides terminal related functions.
package term

import (
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

const defaultWidth = 80

// Info return informations on the terminal associated with the Writer
// argument.
func Info(writer io.Writer) (bool, int, bool) {
	var (
		err   error
		width int
	)

	fwriter, ok := writer.(*os.File)
	if !ok {
		return false, defaultWidth, true
	}

	fdwriter := int(fwriter.Fd())
	if !term.IsTerminal(fdwriter) {
		return false, defaultWidth, true
	}

	width, _, err = term.GetSize(fdwriter)
	if err != nil {
		width = 80
	}

	dark := lipgloss.HasDarkBackground(fwriter, fwriter)

	return true, width, dark
}
