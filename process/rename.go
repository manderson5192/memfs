package process

import (
	"github.com/manderson5192/memfs/filepath"
	"github.com/pkg/errors"
)

func (p *processContext) Rename(srcPath, dstPath string) error {
	// If one path is relative but the other is absolute, then use the working directory to make
	// the relative path into an absolute one.
	baseDir := p.workdir
	srcPathRelative := filepath.Clean(srcPath)
	dstPathRelative := filepath.Clean(dstPath)
	if filepath.IsAbsolutePath(srcPath) && filepath.IsAbsolutePath(dstPath) {
		baseDir = p.fileSystem.RootDirectory()
		// Trim the leading file separators
		srcPathRelative = srcPathRelative[1:]
		dstPathRelative = dstPathRelative[1:]
	} else if filepath.IsAbsolutePath(srcPath) != filepath.IsAbsolutePath(dstPath) {
		// Convert both paths to be absolute
		baseDir = p.fileSystem.RootDirectory()
		workdir, err := p.WorkingDirectory()
		if err != nil {
			return errors.Wrapf(err, "unable to rename %s to %s", srcPath, dstPath)
		}
		if filepath.IsRelativePath(srcPath) {
			srcPathRelative = filepath.Join(workdir, srcPathRelative)
		}
		if filepath.IsRelativePath(dstPath) {
			dstPathRelative = filepath.Join(workdir, dstPathRelative)
		}
		// Trim the leading file separators
		srcPathRelative = srcPathRelative[1:]
		dstPathRelative = dstPathRelative[1:]
	}
	// Do the rename operation
	if err := baseDir.Rename(srcPathRelative, dstPathRelative); err != nil {
		return errors.Wrapf(err, "could not rename %s to %s", srcPath, dstPath)
	}
	return nil
}
