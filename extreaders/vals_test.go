package extreaders_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/stoned/tpkl/extreaders"
)

// TestModuleReaderImplementsPklModuleReaderInterface tests that extreaders.ModuleReader implements pkl.ModuleReader.
func TestValsReaderImplementsPklValsReaderInterface(t *testing.T) {
	var (
		valsReader any
		err        error
	)

	t.Parallel()

	valsReader, err = extreaders.NewValsResourceReader()
	if err != nil {
		t.Fatalf("cannot create a vals resource reader: %s", err)
	}

	if _, ok := valsReader.(pkl.ResourceReader); !ok {
		t.Errorf("expected type %T to implement pkl.ResourceReader", valsReader)
	}
}

const valsExpectedScheme = "vals"

// TestValsScheme test the scheme of the vals external resource reader.
func TestValsScheme(t *testing.T) {
	var (
		valsReader pkl.ResourceReader
		err        error
	)

	t.Parallel()

	valsReader, err = extreaders.NewValsResourceReader()
	if err != nil {
		t.Fatalf("cannot create a vals resource reader: %s", err)
	}

	got := valsReader.Scheme()
	if got != valsExpectedScheme {
		t.Errorf("expected %q, got %q", valsExpectedScheme, got)
	}
}

// TestValsReaderRead test the Read method of vals external resource reader.
func TestValsReaderRead(t *testing.T) {
	t.Parallel()

	valsReader, err := extreaders.NewValsResourceReader()
	if err != nil {
		t.Fatalf("cannot create a vals resource reader: %s", err)
	}

	home, ok := os.LookupEnv("HOME")
	if !ok {
		t.Fatalf("cannot get environment variable HOME")
	}

	testCases := []struct {
		URL      string
		expected string
	}{
		{valsExpectedScheme + ":ref+echo://foo/bar#/foo", "bar"},
		{valsExpectedScheme + ":secretref+echo://foo/bar#/foo", "bar"},
		{valsExpectedScheme + ":ref+echo://foo/bar", "foo/bar"},
		{valsExpectedScheme + ":secretref+echo://foo/bar", "foo/bar"},
		{valsExpectedScheme + ":ref+envsubst://$HOME", home},
		{valsExpectedScheme + ":secretref+envsubst://$HOME", home},
	}

	for _, testCase := range testCases {
		t.Run(testCase.URL, func(t *testing.T) {
			t.Parallel()

			parsedURL, err := url.Parse(testCase.URL)
			if err != nil {
				t.Fatalf("cannot parse URL %q: %s", testCase.URL, err)
			}

			got, err := valsReader.Read(*parsedURL)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			if string(got) != testCase.expected {
				t.Errorf("expected %q, got %q", testCase.expected, got)
			}
		})
	}
}

// TestValsDummy stupidly make us reach more test coverage.
func TestValsDummy(t *testing.T) {
	var reader extreaders.ModuleReader

	t.Parallel()

	dummyURL, _ := url.Parse("")

	if reader.IsGlobbable() {
		t.Error("expected extreaders.ModuleReader.IsGlobbable() to be false")
	}

	if !reader.IsLocal() {
		t.Error("expected extreaders.ModuleReader.IsGlobbable() to be true")
	}

	if reader.HasHierarchicalUris() {
		t.Error("expected extreaders.ModuleReader.HasHierarchicalUris() to be false")
	}

	ary, err := reader.ListElements(*dummyURL)
	if err != nil {
		t.Fatalf("unexpected error %q from extreaders.ModuleReader.ListElements()", err)
	}

	if ary != nil {
		t.Errorf("expected nil return value from extreaders.ModuleReader.ListElements(), got %v", ary)
	}
}
