// Package markdown provides Markdown related functions.
package markdown

import (
	"fmt"
	"io"
	"os"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"github.com/stoned/tpkl/internal/term"
)

// NewRenderer returns a configured Markdown renderer for the provided
// writer, eventually associated to a terminal, and indentation level.
func NewRenderer(writer io.Writer, indent int) (*glamour.TermRenderer, error) {
	isTerminal, termWidth, isDark := term.Info(writer)

	withGlamourStyle := func() glamour.TermRendererOption {
		style := os.Getenv("GLAMOUR_STYLE")
		if style != "" {
			return glamour.WithStylePath(style)
		}

		if !isTerminal {
			return glamour.WithStylePath(styles.NoTTYStyle)
		}

		if isDark {
			return glamour.WithStylePath(styles.DarkStyle)
		}

		return glamour.WithStylePath(styles.LightStyle)
	}

	renderer, err := glamour.NewTermRenderer(
		withGlamourStyle(),
		glamour.WithWordWrap(termWidth-indent),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating markdown renderer: %w", err)
	}

	return renderer, nil
}
