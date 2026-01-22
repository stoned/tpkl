package testscriptcmds_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stoned/tpkl/internal/testscriptcmds"
)

//go:generate go tool txtar -o testdata/script/empty.txtar -c testdata/script/empty/script

func TestTestscriptCmds(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"empty": testscriptcmds.Empty,
		},
	})
}
