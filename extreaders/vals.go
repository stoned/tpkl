package extreaders

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/apple/pkl-go/pkl"
	"github.com/helmfile/vals"
)

const (
	valsCacheSize = 512
)

var (
	// ErrValsEval signals an evaluation error with a vals external resources reader.
	ErrValsEval = errors.New("vals external resource reader evaluation error")
	// ErrValsNew signals an error with a vals external resources reader creation.
	ErrValsNew = errors.New("vals external resource reader creation error")
)

// ValsReader type.
type ValsReader struct {
	scheme  string
	runtime *vals.Runtime
}

// ValsScheme is the vals external resource reader's scheme.
const ValsScheme = "vals"

// NewValsResourceReader returns an initialized ValsReader struct.
func NewValsResourceReader() (*ValsReader, error) {
	runtime, err := vals.New(vals.Options{CacheSize: valsCacheSize})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValsNew, err)
	}

	return &ValsReader{
		scheme:  ValsScheme,
		runtime: runtime,
	}, nil
}

// NewWithValsResourceReader returns a Pkl evaluator options function setter to
// register Vals external resources reader.
func NewWithValsResourceReader() (func(opts *pkl.EvaluatorOptions), error) {
	reader, err := NewValsResourceReader()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValsNew, err)
	}

	return func(opts *pkl.EvaluatorOptions) {
		opts.ResourceReaders = append(opts.ResourceReaders, reader)
		opts.AllowedResources = append(opts.AllowedResources,
			regexp.QuoteMeta(reader.Scheme()+":"))
	}, nil
}

// Scheme returns the scheme of the vals reader.
func (r ValsReader) Scheme() string {
	return r.scheme
}

// IsGlobbable tells if the reader supports globbing.
func (r ValsReader) IsGlobbable() bool {
	return false
}

// IsLocal tells if the resources represented by the reader is considered local to the runtime.
func (r ValsReader) IsLocal() bool {
	return false
}

// HasHierarchicalUris tells if the URIs handled by the reader are hierarchical.
func (r ValsReader) HasHierarchicalUris() bool {
	return false
}

// ListElements returns the list of elements at a specified path.
func (r ValsReader) ListElements(_ url.URL) ([]pkl.PathElement, error) {
	return nil, nil
}

// Read evaluates a vals URL.
func (r ValsReader) Read(uri url.URL) ([]byte, error) {
	// drop the `vals` scheme
	uri.Scheme = ""

	value, err := r.runtime.Get(uri.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValsEval, err)
	}

	return []byte(value), nil
}
