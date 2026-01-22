package enumarg_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	"github.com/stoned/tpkl/internal/enumarg"
)

// TestEnumArgImplementsValueInterface tests that EnumArg implements pflag.Value interface.
func TestEnumArgImplementsValueInterface(t *testing.T) {
	t.Parallel()

	var flag any = new(enumarg.EnumArg)
	if _, ok := flag.(pflag.Value); !ok {
		t.Fatalf("expected type %T to implement pflag.Value", flag)
	}
}

// TestType test EnumArg Type().
func TestType(t *testing.T) {
	t.Parallel()

	flag := enumarg.New([]string{"a", "b"}, "a")
	r := flag.Type()

	if r != "string" {
		t.Errorf("expected enumarg.Type() to return %q, got %q", "string", r)
	}
}

// TestNew test the enumarg constructor.
func TestNew(t *testing.T) {
	t.Parallel()

	want := &enumarg.EnumArg{Allowed: []string{"a", "b"}, Value: "a"}

	got := enumarg.New([]string{"a", "b"}, "a")
	diff := cmp.Diff(want, got)

	if diff != "" {
		t.Errorf("Mismatch (-want, +got):\n%s", diff)
	}
}

// TestAllowed test EnumArg Allowed value constraint.
func TestSet(t *testing.T) {
	t.Parallel()

	cases := []struct {
		set           string
		expectedError error
	}{
		{set: "a"},
		{
			set:           "A",
			expectedError: enumarg.ErrUnsupportedFlagValue,
		},
	}

	for _, testCase := range cases {
		flag := enumarg.New([]string{"a", "b"}, "a")

		err := flag.Set(testCase.set)
		if testCase.expectedError != nil {
			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("expecting error %q, got %q", testCase.expectedError, err)
			}
		} else {
			got := flag.String()
			if got != testCase.set {
				t.Errorf("expecting flag string value %q, got %q", testCase.set, got)
			}
		}
	}
}

// TestDefault test EnumArg default value.
func TestDefault(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flag     *enumarg.EnumArg
		expected string
	}{
		{
			flag:     enumarg.New([]string{"a", "b"}, "a"),
			expected: "a",
		},
		{
			flag:     enumarg.New([]string{"a", "b"}, ""),
			expected: "",
		},
	}

	for _, testCase := range cases {
		got := testCase.flag.String()
		if got != testCase.expected {
			t.Errorf("expecting flag default value %q, got %q", testCase.expected, got)
		}
	}
}
