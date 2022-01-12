package process

import (
	"strings"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/filepath"
	"github.com/pkg/errors"
)

func (p *processContext) MakeDirectory(path string) error {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	if _, err := baseDir.Mkdir(relativePath); err != nil {
		return errors.Wrapf(err, "could not create directory '%s'", path)
	}
	return nil
}

func (p *processContext) ListDirectory(path string) ([]directory.DirectoryEntry, error) {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	entries, err := baseDir.ReadDir(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list entries in directory '%s'", path)
	}
	return entries, nil
}

func (p *processContext) RemoveDirectory(path string) error {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	if err := baseDir.Rmdir(relativePath); err != nil {
		return errors.Wrapf(err, "could not remove directory '%s'", path)
	}
	return nil
}

func (p *processContext) MakeDirectoryWithAncestors(path string) error {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	// Iterate over each part of the path, creating the directory for that part and then looking
	// up the result.  We can ignore errors on directory creation (as would happen if the ancestor
	// directory already existed) so long as the subsequent lookup works
	pathParts := strings.Split(relativePath, filepath.PathSeparator)
	for idx, pathPart := range pathParts {
		var lookupErr error
		_, mkdirErr := baseDir.Mkdir(pathPart)
		baseDir, lookupErr = baseDir.LookupSubdirectory(pathPart)
		if lookupErr != nil {
			errToWrap := mkdirErr
			if errToWrap == nil {
				errToWrap = lookupErr
			}
			ancestor := filepath.Join(pathParts[0 : idx+1]...)
			return errors.Wrapf(errToWrap, "could not create ancestor '%s' of path '%s'", ancestor, path)
		}
	}
	return nil
}
