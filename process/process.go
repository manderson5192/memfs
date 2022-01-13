package process

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/filesys"
)

// ProcessFilesystemContext is an interface that closely resembles the POSIX filesystem interface
// that is available to Linux processes
type ProcessFilesystemContext interface {
	// WorkingDirectory gets the process's current working directory
	WorkingDirectory() (string, error)
	// ChangeDirectory changes the working directory to the specified directory.  Accepts absolute
	// or relative paths.  Returns nil if successful, an error otherwise
	ChangeDirectory(path string) error
	// MakeDirectory creates the specified directory.  Accepts absolute or relative paths.  Returns nil
	// if successful, an error otherwise
	MakeDirectory(dir string) error
	// MakeDirectoryWithAncestors creates the specified path and any ancestor directories that do
	// not already exists.  Unlike MakeDirectory(), this method will not return an error if the
	// specific path is a directory already exists.  Returns an error otherwise
	MakeDirectoryWithAncestors(path string) error
	// ListDirectory returns an array of DirectoryEntry in the specified directory.  Accepts
	// absolute or relative path names.  Returns an array if successful, an error otherwise
	ListDirectory(dir string) ([]directory.DirectoryEntry, error)
	// RemoveDirectory removes the specified directory.  Accepts absolute or relative paths.  Returns
	// nil if successful, an error otherwise
	RemoveDirectory(dir string) error
	// CreateFile creates the specified file and returns a reference to it.  Accepts absolute or
	// relative paths.  Returns nil and an error if unsuccessful.  This call is equivalent to
	// OpenFile(path, O_RDWR|O_CREATE|O_EXCL)
	CreateFile(path string) (file.File, error)
	// OpenFile opens the specified file in the specified mode and returns a reference to it.
	// Accepts absolute or relative paths.  Returns nil and an error if unsuccessful.  It supports
	// the following os, which can be OR'd together (as with open(2) in Linux):
	//	* O_RDONLY: open in read-only mode
	//	* O_WRONLY: open in write-only mode
	//	* O_RDWR: open in read/write mode
	//	* O_CREATE: create the file if it doesn't exist
	//	* O_APPEND: append to the file on each write (as though file.Seek() was used before each write)
	//	* O_TRUNC: if O_WRONLY or O_RDWR then truncat the file to size 0 on open
	//	* O_EXCL: error if O_CREAT and the file exists
	OpenFile(path string, mode int) (file.File, error)
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
	// FindAll walks the subtree rooted at subtreePath, collecting every path for files and
	// directories whose names matche the supplied entry name.  It returns these paths or an error
	FindAll(subtreePath, name string) ([]string, error)
	// FindFirstMatchingFile walks the subtree rooted at subtreePath and returns the path of the
	// first file whose name matches the supplied regex.  Returns the empty string and an error if
	// the regex is invalid, if the underlying Walk() call fails, or if no match is found.
	//
	// regex is evaluated against filenames using a call to Go's regexp.MatchString(regex, filename).
	// This method may return true if regex matches part but not all of filename (i.e. "a.*" is a
	// match for "foobar").  To avoid tricky bugs, clients should make thoughtful use of '^' and '$'
	// in regexes.
	FindFirstMatchingFile(subtreePath string, regex string) (string, error)
}

type processContext struct {
	fileSystem filesys.FileSystem
	workdir    directory.Directory
}

// NewProcessFilesystemContext creates a processContext, which encapsulates a FileSystem, knowledge
// of the current working directoy, and an implementation of the ProcessFilesystemContext interface
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
