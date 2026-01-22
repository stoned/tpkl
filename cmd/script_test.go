package cmd_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stoned/tpkl/cmd"
	"github.com/stoned/tpkl/internal/testscriptcmds"
)

// Generate testscript test scripts
//go:generate go tool txtar -o testdata/script/eval.txtar -c testdata/script/eval/script -p 3 testdata/script/eval/*.pkl testdata/script/eval/*.txt
//go:generate go tool txtar -o testdata/script/eval-no-pkl.txtar -c testdata/script/eval-no-pkl/script -p 3 testdata/script/eval-no-pkl/*.pkl
//go:generate go tool txtar -o testdata/script/eval-no-tpkl.txtar -c testdata/script/eval-no-tpkl/script -p 3 testdata/script/eval-no-tpkl/*.pkl
//go:generate go tool txtar -o testdata/script/help.txtar -c testdata/script/help/script
//go:generate go tool txtar -o testdata/script/list.txtar -c testdata/script/list/script -p 3 testdata/script/list/*.pkl testdata/script/list/*.txt testdata/script/list/*.go
//go:generate go tool txtar -o testdata/script/run-default.txtar -c testdata/script/run-default/script -p 3 testdata/script/run-default/*.pkl
//go:generate go tool txtar -o testdata/script/run-fallback.txtar -c testdata/script/run-fallback/script -p 3 testdata/script/run-fallback/*.pkl
//go:generate go tool txtar -o testdata/script/run-specific.txtar -c testdata/script/run-specific/script -p 3 testdata/script/run-specific/*.pkl
//go:generate go tool txtar -o testdata/script/simple-module.txtar -c testdata/script/simple-module/script -p 3 testdata/script/simple-module/*/*.pkl

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"tpkl": cmd.Main,
	})
}

func TestScript(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"empty": testscriptcmds.Empty,
		},
	})
}
