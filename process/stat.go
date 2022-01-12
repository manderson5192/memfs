package process

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/pkg/errors"
)

func (p *processContext) Stat(path string) (*directory.FileInfo, error) {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	fileInfo, err := baseDir.Stat(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not stat %s", path)
	}
	return fileInfo, nil
}
