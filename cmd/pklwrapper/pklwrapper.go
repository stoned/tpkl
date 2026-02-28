// Package pklwrapper is a wrapper to run pkl command in testscript.
package pklwrapper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Main is the entrypoint to be called via testscript.Main's commands.
func Main() {
	pkl := os.Getenv("PKL_EXEC")
	if pkl == "" {
		pkl = "pkl"
	}

	allArgs := slices.Concat([]string{pkl}, os.Args[1:])
	command := exec.CommandContext(context.Background(), allArgs[0], allArgs[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		ee := (&exec.ExitError{})
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "Error: `%s`: %s\n", strings.Join(allArgs, " "), err)
		os.Exit(1)
	}
}
