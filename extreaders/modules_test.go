package extreaders_test

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/stoned/tpkl/extreaders"
)

// TestModuleReaderImplementsPklModuleReaderInterface tests that extreaders.ModuleReader implements pkl.ModuleReader.
func TestModuleReaderImplementsPklModuleReaderInterface(t *testing.T) {
	t.Parallel()

	var moduleReader any = new(extreaders.ModuleReader)
	if _, ok := moduleReader.(pkl.ModuleReader); !ok {
		t.Errorf("expected type %T to implement pkl.ModuleReader", moduleReader)
	}
}

const moduleExpectedScheme = "tpkl"

// TestModuleScheme test the scheme of the external module reader.
func TestModuleScheme(t *testing.T) {
	var reader extreaders.ModuleReader

	t.Parallel()

	got := reader.Scheme()
	if got != moduleExpectedScheme {
		t.Errorf("expected %q, got %q", moduleExpectedScheme, got)
	}
}

// TestUnknownModule test reading an unknown module is an error.
func TestUnknownModule(t *testing.T) {
	var reader extreaders.ModuleReader

	t.Parallel()

	modurl, _ := url.Parse(moduleExpectedScheme + ":unknown")
	want := extreaders.ErrUnknownModule

	_, err := reader.Read(*modurl)
	if !errors.Is(err, want) {
		t.Errorf("expected error %q, got %q", want, err)
	}
}

// TestTPKLModule test reading the tpkl module.
func TestTPKLModule(t *testing.T) {
	var reader extreaders.ModuleReader

	t.Parallel()

	modurl, _ := url.Parse(moduleExpectedScheme + ":tpkl")

	mod, err := reader.Read(*modurl)
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	if !strings.Contains(mod, "module tpkl.tpkl\n") {
		t.Errorf("expected tpkl module, got %q", mod)
	}
}

// TestModuleDummy stupidly make us reach more test coverage.
func TestModuleDummy(t *testing.T) {
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
