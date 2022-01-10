package process

import (
	"strings"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/filesys"
	"github.com/pkg/errors"
)

type ProcessFilesystemContext interface {
	// WorkingDirectory gets the process's current working directory
	WorkingDirectory() (string, error)
	// ChangeDirectory changes the working directory to the specified directory.  Accepts absolute
	// or relative paths.  Returns nil if successful, an error otherwise
	ChangeDirectory(path string) error
	// MakeDirectory creates the specified directory.  Accepts absolute or relative paths.  Returns nil
	// if successful, an error otherwise
	MakeDirectory(dir string) error
	// ListDirectory returns an array of DirectoryEntry in the specified directory.  Accepts
	// absolute or relative path names.  Returns an array if successful, an error otherwise
	ListDirectory(dir string) ([]directory.DirectoryEntry, error)
	// RemoveDirectory removes the specified directory.  Accepts absolute or relative paths.  Returns
	// nil if successful, an error otherwise
	RemoveDirectory(dir string) error
	// CreateFile creates the specified file and returns a reference to it.  Accepts absolute or
	// relative paths.  Returns nil and an error if unsuccessful
	CreateFile(path string) (file.File, error)
	// OpenFile opens the specified file and returns a reference to it.  Accepts absolute or
	// relative paths.  Returns nil and an error if unsuccessful
	OpenFile(path string) (file.File, error)
	// DeleteFile deletes the specified file.  Accepts absolute or relative paths.  Returns an error
	// if unsuccessful
	DeleteFile(path string) error
}

type processContext struct {
	fileSystem filesys.FileSystem
	workdir    directory.Directory
}

func NewProcessFilesystemContext(fs filesys.FileSystem) ProcessFilesystemContext {
	return &processContext{
		fileSystem: fs,
		workdir:    fs.RootDirectory(),
	}
}

func (p *processContext) WorkingDirectory() (string, error) {
	return p.workdir.ReversePathLookup()
}

// parsePath determines whether `path` is absolute or relative and, if it is absolute, returns
// a new path that is relative to '/' and the directory.Directory for the filesystem root.
// Otherwise, if `path` is relative, then parsePath returns the original path and the
// directory.Directory for the current working directory.
func (p *processContext) parsePath(path string) (string, directory.Directory) {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(path) {
		baseDir = p.fileSystem.RootDirectory()
		path = strings.TrimLeft(path, filepath.PathSeparator)
	}
	return path, baseDir
}

func (p *processContext) ChangeDirectory(path string) error {
	path, baseDir := p.parsePath(path)
	newDir, lookupErr := baseDir.LookupSubdirectory(path)
	if lookupErr != nil {
		return errors.Wrapf(lookupErr, "could not change directories")
	}
	p.workdir = newDir
	return nil
}

func (p *processContext) MakeDirectory(path string) error {
	path, baseDir := p.parsePath(path)
	if _, err := baseDir.Mkdir(path); err != nil {
		return errors.Wrapf(err, "could not create directory '%s'", path)
	}
	return nil
}

func (p *processContext) ListDirectory(path string) ([]directory.DirectoryEntry, error) {
	path, baseDir := p.parsePath(path)
	entries, err := baseDir.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list entries in directory '%s'", path)
	}
	return entries, nil
}

func (p *processContext) RemoveDirectory(path string) error {
	path, baseDir := p.parsePath(path)
	if err := baseDir.Rmdir(path); err != nil {
		return errors.Wrapf(err, "could not remove directory '%s'", path)
	}
	return nil
}

func (p *processContext) CreateFile(path string) (file.File, error) {
	path, baseDir := p.parsePath(path)
	f, err := baseDir.CreateFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create file '%s'", path)
	}
	return f, nil
}

func (p *processContext) OpenFile(path string) (file.File, error) {
	path, baseDir := p.parsePath(path)
	f, err := baseDir.OpenFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file '%s'", path)
	}
	return f, nil
}

func (p *processContext) DeleteFile(path string) error {
	path, baseDir := p.parsePath(path)
	if err := baseDir.DeleteFile(path); err != nil {
		return errors.Wrapf(err, "could not delete file '%s'", path)
	}
	return nil
}

func (p *processContext) Rename(srcPath, dstPath string) error {
	// If one path is relative but the other is absolute, then use the working directory to make
	// the relative path into an absolute one.
	baseDir := p.workdir
	if filepath.IsAbsolutePath(srcPath) && filepath.IsAbsolutePath(dstPath) {
		baseDir = p.fileSystem.RootDirectory()
		srcPath = strings.TrimLeft(srcPath, filepath.PathSeparator)
		dstPath = strings.TrimLeft(dstPath, filepath.PathSeparator)
	} else if filepath.IsAbsolutePath(srcPath) != filepath.IsAbsolutePath(dstPath) {
		// Convert both paths to be absolute
		baseDir = p.fileSystem.RootDirectory()
		workdir, err := p.WorkingDirectory()
		if err != nil {
			return errors.Wrapf(err, "unable to rename %s to %s", srcPath, dstPath)
		}
		if filepath.IsRelativePath(srcPath) {
			srcPath = filepath.Join(workdir, srcPath)
		}
		if filepath.IsRelativePath(dstPath) {
			dstPath = filepath.Join(workdir, dstPath)
		}
	}
	// Do the rename operation
	if err := baseDir.Rename(srcPath, dstPath); err != nil {
		return errors.Wrapf(err, "could not rename %s to %s", srcPath, dstPath)
	}
	return nil
}
