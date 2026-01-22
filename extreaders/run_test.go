package extreaders_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/google/go-cmp/cmp"
	"github.com/stoned/tpkl/extreaders"
)

// TestTPKLModuleExtReader tests the external module reader.
func TestTPKLModuleExtReader(t *testing.T) {
	t.Parallel()

	manager := pkl.NewEvaluatorManager()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}

	main := filepath.Join(filepath.Dir(filename), "../main.go")

	evaluator, err := manager.NewEvaluator(
		t.Context(),
		pkl.PreconfiguredOptions,
		pkl.WithExternalModuleReader(extreaders.ModuleScheme, pkl.ExternalReader{
			Executable: "go",
			Arguments:  []string{"run", main, "readers"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	text := pkl.TextSource(`result = import("` + extreaders.ModuleScheme + `:tpkl")`)

	got, err := evaluator.EvaluateOutputText(t.Context(), text)
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	want := `result {
  argc = 0
  argv {}
}
`
	diff := cmp.Diff(want, got)

	if diff != "" {
		t.Errorf("Mismatch (-want, +got):\n%s", diff)
	}

	err = evaluator.Close()
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	t.Cleanup(func() {
		_ = manager.Close()
	})
}
