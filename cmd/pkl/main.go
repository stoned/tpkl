// Run Pkl's CLI command.
//
// This program is used to generate Go sources for tpkl from Pkl modules
// with the 'go generate' command, via a "go tool" definition in go.mod.
package main

import (
	"github.com/stoned/tpkl/cmd/pklwrapper"
)

func main() {
	pklwrapper.Main()
}
