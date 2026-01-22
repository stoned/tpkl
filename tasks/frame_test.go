package tasks_test

import (
	"reflect"
	"testing"

	"github.com/stoned/tpkl/tasks"
)

func TestNewFrame(t *testing.T) {
	t.Parallel()

	f := tasks.NewFrame()
	if f == nil {
		t.Errorf("failed to allocate a new frame")
	}

	fVarsValue := reflect.ValueOf(f).Elem().FieldByName("vars")
	if fVarsValue.IsZero() {
		t.Errorf("new frame not properly constructed: vars field has zero value")
	}
}

func TestSetVar(t *testing.T) {
	t.Parallel()

	frame := tasks.NewFrame()

	vars := []struct {
		name string
		val  string
	}{
		{"FOO", "value"},
		{"BAR", ""},
	}
	for _, v := range vars {
		frame.SetVar(v.name, v.val)
	}

	merged := frame.Merge()
	for _, v := range vars {
		if v.val != merged[v.name] {
			t.Errorf("expected %q, got %q", v.val, merged[v.name])
		}
	}
}

func TestSetVars(t *testing.T) {
	t.Parallel()

	frame := tasks.NewFrame()

	vars := map[string]string{
		"FOO": "value",
		"BAR": "",
	}

	frame.SetVars(vars)

	merged := frame.Merge()
	for name, val := range vars {
		if val != merged[name] {
			t.Errorf("expected %q, got %q", val, merged[name])
		}
	}
}
