package process

import (
	"fmt"
	"sort"
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

func (p *processContext) Stat(path string) (*directory.FileInfo, error) {
	path, baseDir := p.parsePath(path)
	fileInfo, err := baseDir.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not stat %s", path)
	}
	return fileInfo, nil
}

var SkipDir = fmt.Errorf("skip directory")

// WalkFunc is the type of the function called by Walk to visit each file or directory
//
// The path argument contains the argument to Walk as a prefix.  That is, if Walk is called with
// root argument "dir" and finds a file named "a" in that directory, the walk function will be
// called with argument "dir/a".
//
// The entry argument is a FileInfo for the named path.
//
// The error result returned by the function controls how Walk continues.  If the function returns
// the special value SkipDir, then Walk skips the current directory (path if info.isDir() is true,
// otherwise path's parent directory).  Otherwise, if the function returns a non-nil error, Walk
// stops entirely and returns that error.
//
// The err argument reports an error related to path, signaling that Walk will not walk into that
// directory.  The function can decide how to handle that error; as described earlier, returning
// the error will cause Walk to stop walking the entire tree.
type WalkFunc func(path string, fileInfo *directory.FileInfo, err error) error

// Walk walks the file tree rooted at root, calling fn for each file or directory in the tree,
// including root.
//
// All errors that arise visiting files and directories are filtered by fn: see the WalkFunc
// documentation for details.  In other words, all errors returned by Walk() represent errors that
// originated from a WalkFunc return value, except for SkipDir, which is converted into nil (this
// error is used internally as a sentinel for controlling Walk()'s iteration).
//
// The files are walked in lexical order, which makes the output deterministic.
func (p *processContext) Walk(path string, f WalkFunc) error {
	fileInfo, err := p.Stat(path)
	if err != nil {
		err = f(path, nil, err)
	} else {
		err = p.walk(path, fileInfo, f)
	}
	if err == SkipDir {
		return nil
	}
	return err
}

type byEntry []directory.DirectoryEntry

func (b byEntry) Len() int           { return len(b) }
func (b byEntry) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byEntry) Less(i, j int) bool { return b[i].Name < b[j].Name }

func (p *processContext) walk(path string, fileInfo *directory.FileInfo, f WalkFunc) error {
	// No further recursion on files, so simply call the WalkFunc and return
	if fileInfo.Type != directory.DirectoryType {
		return f(path, fileInfo, nil)
	}
	// Get the entries in the directory
	entries, err := p.ListDirectory(path)
	walkFnErr := f(path, fileInfo, err)
	// Three cases are possible here:
	// 	(1) err is nil and walkFnErr is nil: call walk() on all items under this directory
	//  (2) err is non-nil.  We can't walk this directory, so we must return.  Per the contract
	//		between Walk() and WalkFunc, WalkFunc should receive err and return an error (could be
	//		the same one or a different one) that will be returned by Walk().
	//	(3) err is nil but WalkFunc is non-nil.  WalkFunc could have returned SkipDir, in which case
	//		we don't want to walk() this directory, or it could have returned an error other than
	//		SkipDir, in which case we also don't want to walk() this directory.  So we return
	if err != nil || walkFnErr != nil {
		return walkFnErr
	}
	// Sort the entries lexicographically
	sort.Sort(byEntry(entries))
	// Iterate over the entries in lexicographic order
	for _, entry := range entries {
		// Construct the path for this entry
		newPath := filepath.Join(path, entry.Name)
		// Stat this entry
		fileInfo, err := p.Stat(newPath)
		if err != nil {
			// We couldn't stat() newPath, so we can't walk() newPath.  We have to call WalkFunc and
			// act on the error that it returns:
			//	(1) no error: continue iterating to the next entry in path.
			//	(2) error is SkipDir: we failed to stat() the directory, so we can't walk() newPath
			//		regardless.  Continue iterating to the next entry in path.
			//	(3) error is something other than SkipDir: Walk() needs to be halted and we need to
			//		return this error up the call stack.
			if err := f(newPath, nil, err); err != nil && err != SkipDir {
				return err
			}
		} else {
			err = p.walk(newPath, fileInfo, f)
			if err != nil {
				// walk() returned an error.  Here are the possible interpretations:
				//	(1) err is SkipDir and newPath is a file.  WalkFunc has indicated that it is
				//		time to stop iterating over path's directory.  Percolate the SkipDir up the
				//		call stack.
				//	(2) err is SkipDir and newPath is a directory.  WalkFunc wants to skip newPath's
				//		directory, which we're already done with at this point, so just keep on
				//		iterating.
				//	(3) err is not SkipDir: at some point WalkFunc returned not-SkipDir, which means
				//		that it is time to stop iterating and pass the error up the call stack.
				if fileInfo.Type != directory.DirectoryType || err != SkipDir {
					return err
				}
			}
		}
	}
	return nil
}
