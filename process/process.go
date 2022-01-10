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

func (p *processContext) ChangeDirectory(dir string) error {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(dir) {
		baseDir = p.fileSystem.RootDirectory()
		dir = strings.TrimLeft(dir, filepath.PathSeparator)
	}
	newDir, lookupErr := baseDir.LookupSubdirectory(dir)
	if lookupErr != nil {
		return errors.Wrapf(lookupErr, "could not change directories")
	}
	p.workdir = newDir
	return nil
}

func (p *processContext) MakeDirectory(dir string) error {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(dir) {
		baseDir = p.fileSystem.RootDirectory()
		dir = strings.TrimLeft(dir, filepath.PathSeparator)
	}
	if _, err := baseDir.Mkdir(dir); err != nil {
		return errors.Wrapf(err, "could not create directory '%s'", dir)
	}
	return nil
}

func (p *processContext) ListDirectory(dir string) ([]directory.DirectoryEntry, error) {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(dir) {
		baseDir = p.fileSystem.RootDirectory()
		dir = strings.TrimLeft(dir, filepath.PathSeparator)
	}
	entries, err := baseDir.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list entries in directory '%s'", dir)
	}
	return entries, nil
}

func (p *processContext) RemoveDirectory(dir string) error {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(dir) {
		baseDir = p.fileSystem.RootDirectory()
		dir = strings.TrimLeft(dir, filepath.PathSeparator)
	}
	if err := baseDir.Rmdir(dir); err != nil {
		return errors.Wrapf(err, "could not remove directory '%s'", dir)
	}
	return nil
}

func (p *processContext) CreateFile(filePath string) (file.File, error) {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(filePath) {
		baseDir = p.fileSystem.RootDirectory()
		filePath = strings.TrimLeft(filePath, filepath.PathSeparator)
	}
	f, err := baseDir.CreateFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create file '%s'", filePath)
	}
	return f, nil
}

func (p *processContext) OpenFile(filePath string) (file.File, error) {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(filePath) {
		baseDir = p.fileSystem.RootDirectory()
		filePath = strings.TrimLeft(filePath, filepath.PathSeparator)
	}
	f, err := baseDir.OpenFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file '%s'", filePath)
	}
	return f, nil
}

func (p *processContext) DeleteFile(filePath string) error {
	baseDir := p.workdir
	if filepath.IsAbsolutePath(filePath) {
		baseDir = p.fileSystem.RootDirectory()
		filePath = strings.TrimLeft(filePath, filepath.PathSeparator)
	}
	if err := baseDir.DeleteFile(filePath); err != nil {
		return errors.Wrapf(err, "could not delete file '%s'", filePath)
	}
	return nil
}
