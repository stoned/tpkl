package modules_test

import (
	"testing"

	"github.com/stoned/tpkl/modules"
)

// TestModules tests the modules.Modules() function.
func TestModules(t *testing.T) {
	t.Parallel()

	mods := []string{"tpkl"}

	for i := range mods {
		if _, ok := modules.Modules()[mods[i]]; !ok {
			t.Errorf("Missing embedded module %q", mods[i])
		}
	}
}
