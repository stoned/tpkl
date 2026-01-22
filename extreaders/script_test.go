package extreaders_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stoned/tpkl/cmd"
	"github.com/stoned/tpkl/cmd/pklwrapper"
)

// Generate testscript test scripts
//
//go:generate go tool txtar -o testdata/pkl-test/test.txtar -c testdata/pkl-test/script -p 2 testdata/pkl-test/PklProject testdata/pkl-test/*.pkl
//
//go:generate go tool txtar -o testdata/script/tpkl-eval.txtar -c testdata/script/tpkl-eval/script -p 3 testdata/script/tpkl-eval/*.pkl testdata/script/tpkl-eval/*.txt
//
//go:generate go tool txtar -o testdata/script/tpkl-run.txtar -c testdata/script/tpkl-run/script -p 3 testdata/script/tpkl-run/*.pkl testdata/script/tpkl-run/*.txt

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"pkl-wrapper": pklwrapper.Main,
		"tpkl":        cmd.Main,
	})
}

func TestScript(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}

func TestPkl(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/pkl-test",
	})
}
