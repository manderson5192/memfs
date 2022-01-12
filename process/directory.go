package process

import (
	"github.com/manderson5192/memfs/directory"
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
