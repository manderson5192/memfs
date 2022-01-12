package process

import (
	"github.com/manderson5192/memfs/file"
	"github.com/pkg/errors"
)

func (p *processContext) CreateFile(path string) (file.File, error) {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	f, err := baseDir.CreateFile(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create file '%s'", path)
	}
	return f, nil
}

func (p *processContext) OpenFile(path string) (file.File, error) {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	f, err := baseDir.OpenFile(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file '%s'", path)
	}
	return f, nil
}

func (p *processContext) DeleteFile(path string) error {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	if err := baseDir.DeleteFile(relativePath); err != nil {
		return errors.Wrapf(err, "could not delete file '%s'", path)
	}
	return nil
}
