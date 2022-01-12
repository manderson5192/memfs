package process

import "github.com/pkg/errors"

func (p *processContext) WorkingDirectory() (string, error) {
	return p.workdir.ReversePathLookup()
}

func (p *processContext) ChangeDirectory(path string) error {
	relativePath, baseDir := p.toCleanRelativePathAndBaseDir(path)
	newDir, lookupErr := baseDir.LookupSubdirectory(relativePath)
	if lookupErr != nil {
		return errors.Wrapf(lookupErr, "could not change directories")
	}
	p.workdir = newDir
	return nil
}
