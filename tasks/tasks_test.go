package tasks_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stoned/tpkl/tasks"
)

func writeFile(t *testing.T, filename string, content string) {
	t.Helper()

	_ = os.Mkdir(filepath.Dir(filename), 0o750)

	err := os.WriteFile(filename, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write to file %q: %s", filename, err)
	}
}

const exampleModuleFragment = `
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

func inoperativeModuleTextSource(t *testing.T) string {
	t.Helper()

	return `import "tpkl:tpk"  // typo` + exampleModuleFragment
}

type moduleTestCase struct {
	src             string
	expectedError   error
	expectedMessage string
}

func inoperativeModuleTestCases(t *testing.T) map[string]moduleTestCase {
	t.Helper()

	return map[string]moduleTestCase{
		"EmptyModule": {
			src:           ``,
			expectedError: tasks.ErrEvaluateExpr,
		},
		"NoTask": {
			src:           `foo = "bar"`,
			expectedError: tasks.ErrEvaluateExpr,
		},
		"PklError": {
			src:             inoperativeModuleTextSource(t),
			expectedError:   tasks.ErrEvaluateExpr,
			expectedMessage: `Pkl Error`,
		},
	}
}

// TestModuleTasks_NoTask tests evaluating tasks from inoperative tpkl Pkl modules.
func TestModuleTasksWithInoperativeModule(t *testing.T) {
	t.Parallel()

	for name, testCase := range inoperativeModuleTestCases(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			moduleFile := filepath.Join(t.TempDir(), t.Name()+".pkl")
			writeFile(t, moduleFile, testCase.src)

			_, err := tasks.ModuleTasks(t.Context(), moduleFile)
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("want error %q, got %q", testCase.expectedError, err)
			}

			if testCase.expectedMessage != "" {
				if !strings.Contains(err.Error(), testCase.expectedMessage) {
					t.Errorf("want error message to contains %q, got %q", testCase.expectedMessage, err.Error())
				}
			}
		})
	}
}

// TestRun_NoTask tests running tasks from inoperative tpkl Pkl modules.
func TestRunInoperativeModule(t *testing.T) {
	t.Parallel()

	for name, testCase := range inoperativeModuleTestCases(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			moduleFile := filepath.Join(t.TempDir(), t.Name()+".pkl")
			writeFile(t, moduleFile, testCase.src)

			err := tasks.Run(t.Context(), "no-such-task"+t.Name(), tasks.WithModule(moduleFile))
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("want error %q, got %q", testCase.expectedError, err)
			}

			if testCase.expectedMessage != "" {
				if !strings.Contains(err.Error(), testCase.expectedMessage) {
					t.Errorf("want error message to contains %q, got %q", testCase.expectedMessage, err.Error())
				}
			}
		})
	}
}
