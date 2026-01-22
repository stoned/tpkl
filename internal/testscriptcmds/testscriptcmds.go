// Package testscriptcmds provides commands for github.com/rogpeppe/go-internal scripts
package testscriptcmds

import (
	"github.com/rogpeppe/go-internal/testscript"
)

// Empty implements a testscript command to test if a file is empty.
func Empty(testScript *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		testScript.Fatalf("usage: [!] empty file")
	}

	file := args[0]
	data := testScript.ReadFile(args[0])

	if neg && len(data) == 0 {
		testScript.Fatalf("file %q is empty", file)
	}

	if !neg && len(data) != 0 {
		testScript.Fatalf("file %q is not empty", file)
	}
}
