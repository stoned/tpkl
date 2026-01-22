// Generate txtar archive.
//
// This program is used to generate testscript txtar archives.
//
// This program does its own globbing so it can be run from a Go source directive
// like:
// `//go:generate go run txtar.go -output foo.txtar -comment whatever *.go`
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stoned/tpkl/internal/spath"
	"golang.org/x/tools/txtar"
)

const prog = "txtar"

func makeArchive(output string, comment string, strip int, patterns []string) {
	archive := &txtar.Archive{}

	if comment != "" {
		commentContent, err := os.ReadFile(comment) // #nosec G304
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", prog, err)
			os.Exit(1)
		}

		archive.Comment = commentContent
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", prog, err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "%s: no file matched globbing pattern: %s\n", prog, pattern)
			os.Exit(1)
		}

		for _, filePath := range files {
			var memberName string
			if strip > 0 {
				memberName = spath.StripPath(filePath, strip)
			} else {
				memberName = filePath
			}

			memberContent, err := os.ReadFile(filePath) // #nosec G304
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", prog, err)
				os.Exit(1)
			}

			af := txtar.File{Name: memberName, Data: memberContent}
			archive.Files = append(archive.Files, af)
		}
	}

	p, _ := filepath.Abs(output)
	fmt.Printf("%s: generating %s\n", prog, p)

	err := os.WriteFile(output, txtar.Format(archive), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", prog, err)
		os.Exit(1)
	}
}

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			"  FILE|GLOB...\n\tfile(s), eventually as globbing pattern, to include in archive output\n")

		flag.PrintDefaults()
	}
	output := flag.String("o", "", "txtar file archive output")
	comment := flag.String("c", "", "file to use as comment for archive output")

	stripPath := flag.Int("p", 0,
		"strip the smallest prefix containing N slashes from each file name in archive file marker line")

	flag.Parse()

	if *output == "" {
		fmt.Fprintf(os.Stderr, "%s: missing flag -o\n", prog)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 && *comment == "" {
		fmt.Fprintf(os.Stderr, "%s: nothing to do! no comment and no file to include in archive.\n", prog)
		os.Exit(1)
	}

	makeArchive(*output, *comment, *stripPath, args)
}
