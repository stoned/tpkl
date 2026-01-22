package tasks_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stoned/tpkl/cmd"
	"github.com/stoned/tpkl/internal/testscriptcmds"
)

// Generate testscript test scripts
//go:generate go tool txtar -o testdata/script/calltask.txtar -c testdata/script/calltask/script -p 3 testdata/script/calltask/*.pkl testdata/script/calltask/*.txt
//go:generate go tool txtar -o testdata/script/cmd.txtar -c testdata/script/cmd/script -p 3 testdata/script/cmd/*.pkl testdata/script/cmd/*.txt
//go:generate go tool txtar -o testdata/script/default-vars.txtar -c testdata/script/default-vars/script -p 3 testdata/script/default-vars/*.pkl testdata/script/default-vars/*.txt
//go:generate go tool txtar -o testdata/script/env.txtar -c testdata/script/env/script -p 3 testdata/script/env/*.pkl testdata/script/env/*.txt
//go:generate go tool txtar -o testdata/script/env-var-flag.txtar -c testdata/script/env-var-flag/script -p 3 testdata/script/env-var-flag/*.pkl testdata/script/env-var-flag/*.txt
//go:generate go tool txtar -o testdata/script/expand.txtar -c testdata/script/expand/script -p 3 testdata/script/expand/*.pkl testdata/script/expand/*.txt
//go:generate go tool txtar -o testdata/script/hidden-tasks.txtar -c testdata/script/hidden-tasks/script -p 3 testdata/script/hidden-tasks/*.pkl testdata/script/hidden-tasks/*.txt
//go:generate go tool txtar -o testdata/script/inheritenv.txtar -c testdata/script/inheritenv/script -p 3 testdata/script/inheritenv/*.pkl testdata/script/inheritenv/*.txt
//go:generate go tool txtar -o testdata/script/mustsucceed.txtar -c testdata/script/mustsucceed/script -p 3 testdata/script/mustsucceed/*.pkl
//go:generate go tool txtar -o testdata/script/nocmd.txtar -c testdata/script/nocmd/script -p 3 testdata/script/nocmd/*.pkl
//go:generate go tool txtar -o testdata/script/projectfile.txtar -c testdata/script/projectfile/script -p 3 testdata/script/projectfile/*.pkl
//go:generate go tool txtar -o testdata/script/property-flag.txtar -c testdata/script/property-flag/script -p 3 testdata/script/property-flag/*.pkl
//go:generate go tool txtar -o testdata/script/sh.txtar -c testdata/script/sh/script -p 3 testdata/script/sh/*.pkl testdata/script/sh/*.txt
//go:generate go tool txtar -o testdata/script/task-args.txtar -c testdata/script/task-args/script -p 3 testdata/script/task-args/*.pkl testdata/script/task-args/*.txt
//go:generate go tool txtar -o testdata/script/taskfiles.txtar -c testdata/script/taskfiles/script -p 3 testdata/script/taskfiles/*.pkl testdata/script/taskfiles/*.txt
//go:generate go tool txtar -o testdata/script/taskscycle.txtar -c testdata/script/taskscycle/script -p 3 testdata/script/taskscycle/*.pkl testdata/script/taskscycle/*.txt
//go:generate go tool txtar -o testdata/script/timeout-eval.txtar -c testdata/script/timeout-eval/script -p 3 testdata/script/timeout-eval/*.pkl
//go:generate go tool txtar -o testdata/script/timeout.txtar -c testdata/script/timeout/script -p 3 testdata/script/timeout/*.pkl
//go:generate go tool txtar -o testdata/script/workingdir.txtar -c testdata/script/workingdir/script -p 3 testdata/script/workingdir/*.pkl testdata/script/workingdir/*.txt

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
