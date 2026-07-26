package tasks_test

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stoned/tpkl/tasks"
)

// TestMan test the Man() function.
func TestMan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err    error
		expr   string
		module string
		props  []string
		task   string
	}{
		{
			expr:   `(?m)\ANo documentation for nodoc\n\z`,
			module: "testdata/modules/tasks.pkl",
			task:   "nodoc",
		},
		{
			err:    tasks.ErrUnknownTask,
			module: "testdata/modules/tasks.pkl",
			task:   "condTask",
		},
		{
			expr:   `(?m).*Conditional task\..*`,
			module: "testdata/modules/tasks.pkl",
			props:  []string{"cond=1"},
			task:   "condTask",
		},
		{
			expr:   `(?m).*\*this\*.* neon task\s.*`,
			module: "testdata/modules/tasks.pkl",
			task:   "neon",
		},
	}

	for _, testCase := range cases {
		desc := fmt.Sprintf("module=%s,task=%s,props=%q", testCase.module, testCase.task, testCase.props)
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			opts := []tasks.ManOption{
				tasks.WithModule(testCase.module),
				tasks.WithProperties(testCase.props),
			}

			err := tasks.Man(t.Context(), buf, testCase.task, opts...)
			if testCase.err != nil {
				if !errors.Is(err, testCase.err) {
					t.Fatalf("want error %q, got %q", testCase.err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error %q", err)
				}

				got := buf.String()

				re := regexp.MustCompile(testCase.expr)
				if !re.MatchString(got) {
					t.Errorf("%q does not match %q", got, testCase.expr)
				}
			}
		})
	}
}
