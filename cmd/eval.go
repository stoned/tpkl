package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stoned/tpkl/extreaders"
)

const pklExecVar = "PKL_EXEC"

// EvalCmd returns a cobra command to run tpkl eval command.
func EvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "eval [<options>] <modules>...",
		Short:              "Render Pkl module(s)",
		Long:               "Render Pkl module(s) with the pkl command",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		Run: func(_ *cobra.Command, args []string) {
			var (
				tpkl     string
				lookupOK bool
			)

			pkl := os.Getenv(pklExecVar)
			if pkl == "" {
				pkl = "pkl"
			}

			tpkl, lookupOK = os.LookupEnv("TPKL")
			if !lookupOK {
				tpkl = os.Args[0]
			}

			extReadersRegArgs := extreaders.ExternalReadersRegistrationArgs(tpkl)
			allArgs := slices.Concat([]string{pkl, "eval"}, extReadersRegArgs, args)
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
		},
	}

	helpTemplate := `{{.Long | trimTrailingWhitespaces}}

Usage:
  {{.Use}}

The Pkl command can be specified with the environment variable ` + "`" + pklExecVar +
		"`" + `and defaults to ` + "`pkl`" + `.

For more information, run ` + "`pkl eval --help`" + `.
`
	cmd.SetHelpTemplate(helpTemplate)

	return cmd
}
