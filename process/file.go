package process

import (
	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/os"
	"github.com/pkg/errors"
)

func (p *processContext) OpenFile(path string, mode int) (file.File, error) {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	f, err := baseDir.OpenFile(relativePath, mode)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file '%s'", path)
	}
	return f, nil
}

func (p *processContext) CreateFile(path string) (file.File, error) {
	f, err := p.OpenFile(path, os.OpenFileModeEqualToCreateFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create file '%s'", path)
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
