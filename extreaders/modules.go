package extreaders

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/apple/pkl-go/pkl"
	"github.com/stoned/tpkl/modules"
)

// ModuleReader type.
type ModuleReader struct{}

// ModuleScheme is the external module reader's scheme.
const ModuleScheme = "tpkl"

// ErrUnknownModule signals an unknown embedded Pkl module.
var ErrUnknownModule = errors.New(ModuleScheme + ": unknown Pkl embedded module")

// Scheme returns the scheme of the module reader.
func (r ModuleReader) Scheme() string {
	return ModuleScheme
}

// IsGlobbable tells if the reader supports globbing.
func (r ModuleReader) IsGlobbable() bool {
	return false
}

// IsLocal tells if the resources represented by the reader is considered local to the runtime.
func (r ModuleReader) IsLocal() bool {
	return true
}

// HasHierarchicalUris tells if the URIs handled by the reader are hierarchical.
func (r ModuleReader) HasHierarchicalUris() bool {
	return false
}

// ListElements returns the list of elements at a specified path.
func (r ModuleReader) ListElements(_ url.URL) ([]pkl.PathElement, error) {
	return nil, nil
}

// Read returns the string contents of the module.
func (r ModuleReader) Read(uri url.URL) (string, error) {
	modules := modules.Modules()
	name := uri.Opaque

	if mod, ok := modules[name]; ok {
		return mod, nil
	}

	return "", fmt.Errorf("%w: `%s`'", ErrUnknownModule, name)
}
