package tasks_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/google/go-cmp/cmp"
	"github.com/stoned/tpkl/tasks"
)

// TestTaskModules tests tasks.ModuleTaks().
func TestTaskModules(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err    any
		module string
		msg    string
		nTasks int
		props  []string
	}{
		"EmptyModule": {
			module: "testdata/modules/empty.pkl",
			nTasks: 0,
		},
		"RandomModule": {
			module: "testdata/modules/other.pkl",
			nTasks: 0,
		},
		"NoTask": {
			module: "testdata/modules/notask.pkl",
			nTasks: 0,
		},
		"SomeTasks": {
			module: "testdata/modules/tasks.pkl",
			nTasks: 2,
		},
		"MoreTasks": {
			props:  []string{"cond"},
			module: "testdata/modules/tasks.pkl",
			nTasks: 3,
		},
		"PklError": {
			err:    new(tasks.EvalError),
			module: "testdata/modules/inoperative.pkl",
			msg:    "Pkl Error",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := []func(*pkl.EvaluatorOptions){
				tasks.WithPklProperties(testCase.props),
			}

			tasks, err := tasks.ModuleTasks(t.Context(), testCase.module, opts...)
			if testCase.err != nil {
				if !errors.As(err, &testCase.err) {
					t.Fatalf("want error %q, got %q", testCase.err, err)
				}
			}

			if testCase.msg != "" {
				if !strings.Contains(err.Error(), testCase.msg) {
					t.Errorf("want error message to contains %q, got %q", testCase.msg, err.Error())
				}
			}

			nTasks := len(tasks)

			diff := cmp.Diff(testCase.nTasks, nTasks)
			if diff != "" {
				t.Errorf("Mismatch in tasks number (-want, +got):\n%s", diff)
			}
		})
	}
}

// TestRunNoTask tests running tasks from inoperative tpkl Pkl modules.
func TestRunNoTask(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		module string
		err    any
		msg    string
	}{
		"EmptyModule": {
			module: "testdata/modules/empty.pkl",
			err:    tasks.ErrUnknownTask,
		},
		"NoTask": {
			module: "testdata/modules/notask.pkl",
			err:    tasks.ErrUnknownTask,
		},
		"NoSuchTask": {
			module: "testdata/modules/tasks.pkl",
			err:    tasks.ErrUnknownTask,
		},
		"PklError": {
			module: "testdata/modules/inoperative.pkl",
			err:    new(tasks.EvalError),
			msg:    `Pkl Error`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tasks.Run(t.Context(), "no-such-task"+t.Name(), tasks.WithModule(testCase.module))
			if testCase.err != nil {
				if !errors.As(err, &testCase.err) {
					t.Fatalf("want error %q, got %q", testCase.err, err)
				}
			}

			if testCase.msg != "" {
				if !strings.Contains(err.Error(), testCase.msg) {
					t.Errorf("want error message to contains %q, got %q", testCase.msg, err.Error())
				}
			}
		})
	}
}

// TestEvalErrorImplementsErrorInterface tests that tasks.EvalError implements error.
func TestEvalErrorImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err any = new(tasks.EvalError)
	if _, ok := err.(error); !ok {
		t.Errorf("expected type %T to implement error", err)
	}
}
