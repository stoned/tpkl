// Package enumarg implements a github.com/spf13/pflag's Value interface for "enumerated flags"
package enumarg

import (
	"errors"
	"fmt"
	"slices"
)

// ErrUnsupportedFlagValue signals an unsupporteg flag value.
var ErrUnsupportedFlagValue = errors.New("unsupported flag value")

// EnumArg is a "enumerated flag" type.
type EnumArg struct {
	Allowed []string
	Value   string
}

// New initializes a EnumArg struct.
func New(allowed []string, d string) *EnumArg {
	return &EnumArg{
		Allowed: allowed,
		Value:   d,
	}
}

func (e *EnumArg) String() string {
	return e.Value
}

// Set an EnumArg value.
func (e *EnumArg) Set(value string) error {
	if !slices.Contains(e.Allowed, value) {
		return fmt.Errorf("%w: %q is not included in %q", ErrUnsupportedFlagValue, value, e.Allowed)
	}

	e.Value = value

	return nil
}

// Type returns the "pflag type" of an "enumerated flag".
func (e *EnumArg) Type() string {
	return "string"
}
