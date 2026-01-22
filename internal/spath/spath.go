// Package spath provides supplemental path functions.
package spath

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// SplitPath returns a cleaned up path's components as a slice.
func SplitPath(path string) []string {
	return strings.Split(filepath.Clean(path), string(filepath.Separator))
}

// StripPath strips num components from path and returns
// a cleaned up path.
// If num is less than 1 it returns the path *unchanged*.
// If num is greater than the number of path's component, an empty string
// is returned.
func StripPath(path string, num int) string {
	if num < 1 {
		return path
	}

	components := SplitPath(path)
	if num > len(components) {
		return ""
	}

	return filepath.Join(components[num:]...)
}

// ErrFindUp signals an error while using FindUp().
var ErrFindUp = errors.New("cannot find file")

// FindUp returns the pathname of a file specified as a relative path,
// searching from 'workingDir' directory up to the 'rootFs' filesystem root.
func FindUp(part string, workingDir string, rootFs fs.FS) (string, error) {
	if filepath.IsAbs(part) {
		return "", fmt.Errorf("%w: `%s`: path is absolute", ErrFindUp, part)
	}

	if !filepath.IsAbs(workingDir) {
		return "", fmt.Errorf("%w: `%s`: current directory is not absolute", ErrFindUp, workingDir)
	}

	var (
		file           fs.File
		err            error
		path, previous string
	)

	base := workingDir
	path = filepath.Join(base, part)

	for path != previous {
		file, err = rootFs.Open(path[1:])
		if err == nil {
			_ = file.Close()

			return path, nil
		}

		previous = path
		base = filepath.Join(base, "..")
		path = filepath.Join(base, part)
	}

	return "", fmt.Errorf("%w: %s", ErrFindUp, part)
}
