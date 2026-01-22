package modules_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stoned/tpkl/cmd"
	"github.com/stoned/tpkl/cmd/pklwrapper"
)

// Generate testscript test scripts
//go:generate go tool txtar -o testdata/script/cmdortask.txtar -c testdata/script/cmdortask/script -p 3 testdata/script/cmdortask/*.pkl
//go:generate go tool txtar -o testdata/script/taskname.txtar -c testdata/script/taskname/script -p 3 testdata/script/taskname/*.pkl
//go:generate go tool txtar -o testdata/pkl-test/test.txtar -c testdata/pkl-test/script -p 2 testdata/pkl-test/PklProject testdata/pkl-test/*.pkl testdata/pkl-test/*.pcf

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
