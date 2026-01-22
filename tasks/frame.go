package tasks

import (
	"maps"
	"os"
	"strings"
	"sync"
)

// Frame represent a task's environment.
type Frame struct {
	vars      map[string]string
	environ   map[string]string
	enclosing *Frame
	merged    map[string]string
	lock      sync.Mutex
}

// NewFrame creates an empty task Frame.
func NewFrame() *Frame {
	frame := &Frame{}
	frame.vars = make(map[string]string)

	return frame
}

// NewEnclosedFrame creates an task Frame with an enclosing frame.
func NewEnclosedFrame(enclosingFrame *Frame) *Frame {
	frame := NewFrame()
	frame.enclosing = enclosingFrame

	return frame
}

// SetEnviron provides access to environment variables from the frame.
func (frame *Frame) SetEnviron() {
	frame.lock.Lock()
	defer frame.lock.Unlock()

	frame.environ = GetEnviron()
	frame.merged = nil
}

// SetVar sets a variable's value in a Frame.
func (frame *Frame) SetVar(name string, value string) {
	frame.lock.Lock()
	defer frame.lock.Unlock()

	frame.vars[name] = value
	frame.merged = nil
}

// SetVars sets variables in a Frame from a map.
func (frame *Frame) SetVars(m map[string]string) {
	frame.lock.Lock()
	defer frame.lock.Unlock()

	maps.Copy(frame.vars, m)
	frame.merged = nil
}

// Merge computes and returns all frame variables in a single map: its own variables,
// inherited variables from enclosing frame, and environment variables.
func (frame *Frame) Merge() map[string]string {
	frame.lock.Lock()
	defer frame.lock.Unlock()

	if frame.merged != nil {
		return frame.merged
	}

	frame.merged = make(map[string]string)

	if frame.environ != nil {
		maps.Copy(frame.merged, frame.environ)
	}

	if frame.enclosing != nil {
		parent := frame.enclosing.Merge()
		maps.Copy(frame.merged, parent)
	}

	maps.Copy(frame.merged, frame.vars)

	return frame.merged
}

// EnvList returns the frame merged variables as list of NAME=VALUE strings.
func (frame *Frame) EnvList() []string {
	merged := frame.Merge()
	env := make([]string, 0, len(merged))

	for name, val := range merged {
		env = append(env, name+"="+val)
	}

	return env
}

// ExpandMapping is a helper function for expansion.Expand().
func (frame *Frame) ExpandMapping() func(string) string {
	merged := frame.Merge()

	return func(input string) string {
		if val, ok := merged[input]; ok {
			return val
		}

		return "$(" + input + ")"
	}
}

var (
	_environOnce sync.Once
	_osEnviron   map[string]string
)

// GetEnviron produces a single copy of environment variables as a map.
func GetEnviron() map[string]string {
	_environOnce.Do(func() {
		env := os.Environ()
		_osEnviron = make(map[string]string, len(env))

		for _, s := range env {
			name, val, _ := strings.Cut(s, "=")
			_osEnviron[name] = val
		}
	})

	return _osEnviron
}
