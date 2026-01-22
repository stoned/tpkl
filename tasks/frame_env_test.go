package tasks_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stoned/tpkl/tasks"
)

func setupFrames(t *testing.T, withEnv bool) (*tasks.Frame, []struct {
	name string
	val  string
},
) {
	t.Helper()

	frame0 := tasks.NewFrame()
	frame1 := tasks.NewEnclosedFrame(frame0)
	frame2 := tasks.NewEnclosedFrame(frame1)

	envVars := []struct {
		name string
		val  string
	}{
		{"TPKL_TEST_VAR_ENV_0", ""},
		{"TPKL_TEST_VAR_ENV_1", "valenv1"},
	}

	if withEnv {
		for _, v := range envVars {
			t.Setenv(v.name, v.val)
		}

		frame1.SetEnviron()
	}

	vars := []struct {
		frame *tasks.Frame
		name  string
		val   string
	}{
		{frame0, "VAR_0", "val0"},
		{frame0, "VAR_B", "val0"},
		{frame1, "VAR_1", "val1"},
		{frame1, "VAR_A", "val1"},
		{frame1, "VAR_B", "val1"},
		{frame1, "VAR_C", "val1"},
		{frame2, "VAR_2", "val2"},
		{frame2, "VAR_A", "val2"},
		{frame2, "VAR_C", ""},
	}
	for _, v := range vars {
		v.frame.SetVar(v.name, v.val)
	}

	expected := []struct {
		name string
		val  string
	}{
		{"VAR_0", "val0"},
		{"VAR_1", "val1"},
		{"VAR_2", "val2"},
		{"VAR_B", "val1"},
		{"VAR_A", "val2"},
		{"VAR_C", ""},
	}

	if withEnv {
		expected = append(expected, envVars...)
	}

	return frame2, expected
}

func TestFrames(t *testing.T) {
	frame, expected := setupFrames(t, false)

	merged := frame.Merge()
	for _, v := range expected {
		if v.val != merged[v.name] {
			diff := cmp.Diff(v.val, merged[v.name])
			t.Errorf("Frames merge mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestFramesWithEnv(t *testing.T) {
	frame, expected := setupFrames(t, true)

	merged := frame.Merge()
	for _, v := range expected {
		if v.val != merged[v.name] {
			diff := cmp.Diff(v.val, merged[v.name])
			t.Errorf("Frames with environment merge mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestExpandMapping(t *testing.T) {
	frame, expectedPresent := setupFrames(t, true)
	expectedAbsent := []struct {
		name string
		val  string
	}{
		{"VAR_NOT_THERE", "$(VAR_NOT_THERE)"},
	}

	mapping := frame.ExpandMapping()

	for _, v := range expectedPresent {
		val := mapping(v.name)
		if val != v.val {
			diff := cmp.Diff(v.val, val)
			t.Errorf("Frames expand mapping mismatch (-want +got):\n%s", diff)
		}
	}

	for _, v := range expectedAbsent {
		val := mapping(v.name)
		if val != v.val {
			diff := cmp.Diff(v.val, val)
			t.Errorf("Frames expand mapping mismatch (-want +got):\n%s", diff)
		}
	}
}
