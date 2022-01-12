package process

import (
	"fmt"
	"regexp"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/pkg/errors"
)

func (p *processContext) FindAll(subtreePath, name string) ([]string, error) {
	paths := make([]string, 0)
	walkFunc := func(path string, fileInfo *directory.FileInfo, err error) error {
		pathInfo := filepath.ParsePath(path)
		if pathInfo.Entry == name {
			paths = append(paths, path)
		}
		return nil
	}
	if err := p.Walk(subtreePath, walkFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to find all files and directories named '%s'", name)
	}
	return paths, nil
}

func (p *processContext) FindFirstMatchingFile(subtreePath string, regex string) (string, error) {
	matchingPath := ""
	matchFound := false
	walkFunc := func(path string, fileInfo *directory.FileInfo, err error) error {
		if fileInfo == nil {
			return fmt.Errorf("unable to determine if %s is a file", path)
		}
		if matchFound {
			// Skip everything once our match has been found
			return SkipDir
		}
		pathInfo := filepath.ParsePath(path)
		matches, err := regexp.MatchString(regex, pathInfo.Entry)
		if err != nil {
			// Propagate regex errors to the return value of Walk()
			return err
		}
		if !matches {
			// Keep Walk()'ing
			return nil
		}
		if fileInfo.Type == directory.FileType {
			// The name matched on a file.  Record the matching path and begin returning SkipDir to
			// successfully terminate Walk() as soon as possible
			matchFound = true
			matchingPath = path
			return SkipDir
		}
		// otherwise, keep Walk()'ing
		return nil
	}
	if err := p.Walk(subtreePath, walkFunc); err != nil {
		return "", errors.Wrapf(err, "unable to find first file matching '%s' under '%s'", regex, subtreePath)
	}
	if !matchFound {
		return "", errors.Wrapf(fserrors.ENoEnt, "no match found")
	}
	return matchingPath, nil
}
