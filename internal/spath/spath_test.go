package spath_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/stoned/tpkl/internal/spath"
)

// TestSplitPath tests SplitPath function.
func TestSplitPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path     string
		expected []string
	}{
		{path: "", expected: []string{"."}},
		{path: "foo", expected: []string{"foo"}},
		{path: "foo/", expected: []string{"foo"}},
		{path: "foo/bar", expected: []string{"foo", "bar"}},
		{path: "foo/bar/baz", expected: []string{"foo", "bar", "baz"}},
		{path: "/foo/bar/baz", expected: []string{"", "foo", "bar", "baz"}},
		{path: "/", expected: []string{"", ""}},
		{path: "/foo/", expected: []string{"", "foo"}},
		{path: "/foo/bar/", expected: []string{"", "foo", "bar"}},
		{path: "/foo///bar/", expected: []string{"", "foo", "bar"}},
		{path: "/foo/bar/../", expected: []string{"", "foo"}},
		{path: "/foo/bar/../../", expected: []string{"", ""}},
	}

	for _, testCase := range cases {
		t.Run(testCase.path, func(t *testing.T) {
			t.Parallel()

			v := spath.SplitPath(testCase.path)
			if diff := cmp.Diff(testCase.expected, v); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestStripath tests StripPath function.
func TestStripPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path     string
		num      int
		expected string
	}{
		{path: "", num: 10, expected: ""},
		{path: "foo", num: 0, expected: "foo"},
		{path: "foo", num: 2, expected: ""},
		{path: "foo/", num: 2, expected: ""},
		{path: "foo/bar", num: 1, expected: "bar"},
		{path: "foo/bar/baz", num: 1, expected: "bar/baz"},
		{path: "foo/bar/baz", num: 2, expected: "baz"},
		{path: "foo/bar/baz", num: 3, expected: ""},
		{path: "/foo/bar/baz", num: 0, expected: "/foo/bar/baz"},
		{path: "/", num: 0, expected: "/"},
		{path: "/", num: 1, expected: ""},
		{path: "/", num: 2, expected: ""},
		{path: "/foo/", num: 0, expected: "/foo/"},
		{path: "/foo/bar/", num: 0, expected: "/foo/bar/"},
		{path: "/foo///bar/", num: 1, expected: "foo/bar"},
		{path: "/foo/bar/../", num: 1, expected: "foo"},
		{path: "/foo/bar/../../", num: 1, expected: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.path+"["+strconv.Itoa(testCase.num)+"]", func(t *testing.T) {
			t.Parallel()

			v := spath.StripPath(testCase.path, testCase.num)
			if diff := cmp.Diff(testCase.expected, v); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFindUp tests FindUp function.
func TestFindUp(t *testing.T) {
	t.Parallel()

	fsTest := fstest.MapFS{
		"home/user/work/tasks.pkl":    {},
		"home/user/work/dir1/foo.txt": {},
		"home/user/work/dir2/foo.txt": {},
		"top.txt":                     {},
	}

	cases := []struct {
		part          string
		wd            string
		expected      string
		expectedError error
	}{
		{"tasks.pkl", "/home/user/work", "/home/user/work/tasks.pkl", nil},
		{"top.txt", "/home/user/work", "/top.txt", nil},
		{"work/tasks.pkl", "/home/user/work", "/home/user/work/tasks.pkl", nil},
		{"dir1/foo.txt", "/home/user/work/dir1", "/home/user/work/dir1/foo.txt", nil},
		{"dir1/foo.txt", "/home/user/work/dir2", "/home/user/work/dir1/foo.txt", nil},
		{"foo.txt", "/home/user/work/dir1", "/home/user/work/dir1/foo.txt", nil},
		{"foo.txt", "/home/user/work/dir2", "/home/user/work/dir2/foo.txt", nil},
		{"bar.txt", "/home/user/work/dir2", "", spath.ErrFindUp},
		{"bar.txt", "/", "", spath.ErrFindUp},
		{"top.txt", "/", "/top.txt", nil},
		{"bar.txt", "", "", spath.ErrFindUp},
		{"bar.txt", ".", "", spath.ErrFindUp},
		{"bar.txt", ".", "", spath.ErrFindUp},
		{"bar.txt", "dir", "", spath.ErrFindUp},
		{"/foo", "/", "", spath.ErrFindUp},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("part=%q wd=%q", testCase.part, testCase.wd), func(t *testing.T) {
			t.Parallel()

			path, err := spath.FindUp(testCase.part, testCase.wd, fsTest)
			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					t.Errorf("expected error %v, got error: %v", testCase.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if path != testCase.expected {
					t.Errorf("expected %q, got %q", testCase.expected, path)
				}
			}
		})
	}
}
