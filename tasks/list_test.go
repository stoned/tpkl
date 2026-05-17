package tasks_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stoned/tpkl/tasks"
)

//go:embed testdata/modules/*.json
var jsonFiles embed.FS

func jsonDocument(t *testing.T, name string) string {
	t.Helper()

	content, err := jsonFiles.ReadFile(name)
	if err != nil {
		t.Fatalf("failed to get embedded file %q: %s", name, err)
	}

	return string(content)
}

// TestListUnsupportedFormat the List() function with an unsupported format.
func TestListUnsupportedFormat(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	want := tasks.ErrUnknownOption

	err := tasks.List(t.Context(), b, "uNsuPpoRtedOptIOn", tasks.WithModule("testdata/modules/tasks.pkl"))
	if !errors.Is(err, want) {
		t.Errorf("expecting error %q, got %q", want, err)
	}
}

// TestListName test the List() function with the "name" format.
func TestListName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		module   string
		props    []string
		expected string
	}{
		{
			module:   "testdata/modules/notask.pkl",
			expected: "",
		},
		{
			module:   "testdata/modules/abc.pkl",
			expected: "a\nb\nc\n",
		},
		{
			module:   "testdata/modules/acb.pkl",
			expected: "a\nb\nc\n",
		},
		{
			module:   "testdata/modules/tasks.pkl",
			expected: "neon\nnodoc\ntask\n",
		},
		{
			module:   "testdata/modules/tasks.pkl",
			props:    []string{"cond=1"},
			expected: "condTask\nneon\nnodoc\ntask\n",
		},
	}

	for _, testCase := range cases {
		desc := fmt.Sprintf("module=%s,props=%q", testCase.module, testCase.props)
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			opts := []tasks.ListOption{
				tasks.WithModule(testCase.module),
				tasks.WithProperties(testCase.props),
			}

			err := tasks.List(t.Context(), buf, "name", opts...)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			got := buf.String()

			diff := cmp.Diff(testCase.expected, got)
			if diff != "" {
				t.Errorf("Mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}

// TestListJSON test the List() function with the "json" format.
func TestListJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		module   string
		props    []string
		expected string
	}{
		{
			module:   "testdata/modules/notask.pkl",
			expected: jsonDocument(t, "testdata/modules/notask.json"),
		},
		{
			module:   "testdata/modules/abc.pkl",
			expected: jsonDocument(t, "testdata/modules/abc.json"),
		},
		{
			module:   "testdata/modules/acb.pkl",
			expected: jsonDocument(t, "testdata/modules/abc.json"),
		},
		{
			module:   "testdata/modules/tasks.pkl",
			props:    []string{"cond=1"},
			expected: jsonDocument(t, "testdata/modules/tasks-cond.json"),
		},
		{
			module:   "testdata/modules/tasks.pkl",
			expected: jsonDocument(t, "testdata/modules/tasks.json"),
		},
	}

	jsonTransformer := cmpopts.AcyclicTransformer("unmarshalJSON", func(s string) map[string]any {
		var ret map[string]any

		err := json.Unmarshal([]byte(s), &ret)
		if err != nil {
			t.Fatalf("unexpected error decoding JSON %v: %q", s, err)
		}

		return ret
	})

	for _, testCase := range cases {
		desc := fmt.Sprintf("module=%s,props=%q", testCase.module, testCase.props)
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			opts := []tasks.ListOption{
				tasks.WithModule(testCase.module),
				tasks.WithProperties(testCase.props),
			}

			err := tasks.List(t.Context(), buf, "json", opts...)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			diff := cmp.Diff(testCase.expected, buf.String(), jsonTransformer)
			if diff != "" {
				t.Errorf("Mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}

type reCase struct {
	re   *regexp.Regexp
	expr string
}

// TestListSummary test the List() function with the "summary" format.
func TestListSummary(t *testing.T) {
	t.Parallel()

	newReCase := func(e string) reCase {
		return reCase{
			re:   regexp.MustCompile(e),
			expr: e,
		}
	}

	cases := []struct {
		reCase

		module string
		props  []string
	}{
		{
			module: "testdata/modules/abc.pkl",
			reCase: newReCase(`(?m)\Aa\nb\nc\n\z`),
		},
		{
			module: "testdata/modules/notask.pkl",
			reCase: newReCase(`(?m)\A\z`),
		},
		{
			module: "testdata/modules/tasks.pkl",
			props:  []string{"cond"},
			reCase: newReCase(`(?m)\A` +
				`condTask\s+Conditional task\.\s+\n` +
				`neon\s+\*this\* \*\*is\*\* \*the\* neon task\s+\n` +
				`nodoc\s+\n` +
				`task\s+The \*\*task\*\* task\.\s+\n\z`),
		},
		{
			module: "testdata/modules/tasks.pkl",
			reCase: newReCase(`(?m)\A` +
				`neon\s+\*this\* \*\*is\*\* \*the\* neon task\s+\n` +
				`nodoc\s+\n` +
				`task\s+The \*\*task\*\* task\.\s+\n\z`),
		},
	}

	for _, testCase := range cases {
		desc := fmt.Sprintf("module=%s,props=%q", testCase.module, testCase.props)
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			opts := []tasks.ListOption{
				tasks.WithModule(testCase.module),
				tasks.WithProperties(testCase.props),
			}

			err := tasks.List(t.Context(), buf, "summary", opts...)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			got := buf.String()
			if !testCase.re.MatchString(got) {
				t.Errorf("%q does not match %q", got, testCase.expr)
			}
		})
	}
}
