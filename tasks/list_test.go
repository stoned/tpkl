package tasks_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stoned/tpkl/tasks"
)

func testModule(t *testing.T) string {
	t.Helper()

	moduleSrc := `
import "tpkl:tpkl"
tasks: tpkl.Tasks = new {
  tasks {
    ["c"] {
      cmds {
        new {
          cmd { "echo 'Hello, world!'" }
        }
      }
    }
  }
}
`
	moduleFile := filepath.Join(t.TempDir(), "module.pkl")
	_ = os.Mkdir(filepath.Dir(moduleFile), 0o750)

	err := os.WriteFile(moduleFile, []byte(moduleSrc), 0o600)
	if err != nil {
		t.Fatalf("failed to write to file %q: %s", moduleFile, err)
	}

	return moduleFile
}

// TestListUnsupportedFormat the List() function with an unsupported format.
func TestListUnsupportedFormat(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	module := testModule(t)
	want := tasks.ErrUnknownOption

	err := tasks.List(t.Context(), b, "uNsuPpoRtedOptIOn", tasks.WithModule(module))
	if !errors.Is(err, want) {
		t.Errorf("expecting error %q, got %q", want, err)
	}
}

// TestListName test the List() function with the "name" option.
func TestListName(t *testing.T) {
	t.Parallel()

	buff := new(bytes.Buffer)
	module := testModule(t)

	err := tasks.List(t.Context(), buff, "name", tasks.WithModule(module))
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	want := "c\n"
	got := buff.String()

	diff := cmp.Diff(want, got)
	if diff != "" {
		t.Errorf("Mismatch (-want, +got):\n%s", diff)
	}
}

// TestListJSON test the List() function with the "json" option.
func TestListJSON(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	module := testModule(t)

	err := tasks.List(t.Context(), buf, "json", tasks.WithModule(module))
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	var d any

	err = json.Unmarshal(buf.Bytes(), &d)
	if err != nil {
		t.Errorf("unexpected error decoding JSON %q", err)
	}
}
