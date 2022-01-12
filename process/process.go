package process

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/filesys"
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
	// Rename moves the file or directory at srcPath to dstPath.  If dstPath already exists, then
	// it will attempt to remove that file or directory.  Returns an error if unsuccessful.
	Rename(srcPath, dstPath string) error
	// Stat returns a file.FileInfo for the specified file or directory, or an error.
	Stat(path string) (*directory.FileInfo, error)
	// Walk walks the file tree rooted at root, calling fn for each file or directory in the tree,
	// including root.
	//
	// All errors that arise visiting files and directories are filtered by fn: see the WalkFunc
	// documentation for details.  In other words, all errors returned by Walk() represent errors that
	// originated from a WalkFunc return value, except for SkipDir, which is converted into nil (this
	// error is used internally as a sentinel for controlling Walk()'s iteration).
	//
	// The files are walked in lexical order, which makes the output deterministic.
	Walk(path string, f WalkFunc) error
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

// toCleanRelativePathAndBaseDir examines whether path is absolute or relative and, based on that
// insight, returns a base directory (either the root directory or the working directory) and a
// relative (to the base directory) path that is equivalent to path.  It also uses filepath.Path()
// to cleanup path before examination.
func (p *processContext) toCleanRelativePathAndBaseDir(path string) (string, directory.Directory) {
	baseDir := p.workdir
	path = filepath.Clean(path)
	if filepath.IsAbsolutePath(path) {
		baseDir = p.fileSystem.RootDirectory()
		// Trim the leading file separator
		path = path[1:]
	}
	return path, baseDir
}
