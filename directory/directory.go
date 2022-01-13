package directory

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/inode"
	"github.com/manderson5192/memfs/modes"
	"github.com/pkg/errors"
)

const (
	SelfDirectoryEntry   = filepath.SelfDirectoryEntry
	ParentDirectoryEntry = filepath.ParentDirectoryEntry
)

type DirectoryEntryType int

func directoryEntryTypeFromInodeType(t inode.InodeType) DirectoryEntryType {
	if t == inode.InodeDirectory {
		return DirectoryType
	} else if t == inode.InodeFile {
		return FileType
	} else {
		return InvalidType
	}
}

const (
	InvalidType DirectoryEntryType = iota
	DirectoryType
	FileType
)

func (t DirectoryEntryType) MarshalJSON() ([]byte, error) {
	toReturn := "invalid"
	switch t {
	case DirectoryType:
		toReturn = "directory"
	case FileType:
		toReturn = "file"
	default:
		toReturn = "invalid"
	}
	return json.Marshal(toReturn)
}

// DirectoryEntry represents a file or directory entry in a given directory
type DirectoryEntry struct {
	// Name is the entry's name
	Name string `json:"name"`
	// Type indicates whether the entry is a file or a directory
	Type DirectoryEntryType `json:"type"`
}

// FileInfo represents information about a single file or directory.  If Type indicates a directory,
// then Size will be the number of directory entries.  If Type indicates a file, then Size will be
// the file's size in bytes
type FileInfo struct {
	Size int
	Type DirectoryEntryType
}

type Directory interface {
	// Equals returns true if the other Directory references the same inode, false otherwise
	Equals(other Directory) bool
	// ReversePathLookup returns a valid absolute path for the directory or an error
	ReversePathLookup() (string, error)
	// LookupSubdirectory returns the Directory for the subdirectory of the current directory, or an
	// error.  If subdirectory is empty, then this Directory itself will be returned.
	LookupSubdirectory(subdirectory string) (Directory, error)
	// Mkdir creates and returns a Directory for the specified subdirectory of the current
	// directory, or returns an error.  It will return an error if a path component does not exist
	// or is not a directory.  It will return an error if the specified subdirectory already exists.
	Mkdir(subdirectory string) (Directory, error)
	// ReadDir returns an array of DirectoryEntry for the specified subdirectory of the current
	// directory, or returns an error.  It will return an error if a path component does not exist
	// or is not a directory.
	ReadDir(subdirectory string) ([]DirectoryEntry, error)
	// Rmdir removes the specified subdirectory of the current directory, or returns an error
	Rmdir(subdirectory string) error
	// CreateFile creates a new file at the specified relative path, or returns an error
	CreateFile(relativePath string) (file.File, error)
	// OpenFile returns a reference to the specified relative path in the specified mode, or returns
	// an error
	OpenFile(relativePath string, mode int) (file.File, error)
	// DeleteFile removes the specified file, which must be at a path relative to the current
	// directory.  It returns an error if it is unsuccessful
	DeleteFile(relativePath string) error
	// Rename moves the file or directory at the specified relative src path to the specified
	// relative dst path.  If an entry already exists at the dst path, then this operation will
	// attempt to atomically replace it.  Returns an error if unsuccessful
	Rename(srcPath, dstPath string) error
	// Stat returns a FileInfo for the file or directory at the indicated path.  If relativePath is
	// empty, then the indicated path will for the receiver Directory object
	Stat(relativePath string) (*FileInfo, error)
}

type directory struct {
	*inode.DirectoryInode
}

func NewDirectory(inode *inode.DirectoryInode) Directory {
	return &directory{
		DirectoryInode: inode,
	}
}

func (d *directory) Equals(other Directory) bool {
	if d == nil || other == nil {
		return false
	}
	otherDir, ok := other.(*directory)
	if !ok {
		return false
	}
	return d.DirectoryInode == otherDir.DirectoryInode
}

func (d *directory) ReversePathLookup() (string, error) {
	pathParts := []string{}
	currentDirInode := d.DirectoryInode
	for !currentDirInode.IsRootDirectoryInode() {
		parentDirInode := currentDirInode.Parent()
		pathPart, err := parentDirInode.ReverseLookupEntry(currentDirInode)
		if err != nil {
			return "", errors.Wrapf(err, "could not complete reverse path lookup")
		}
		pathParts = append([]string{pathPart}, pathParts...)
		currentDirInode = parentDirInode
	}
	path := strings.Join(pathParts, filepath.PathSeparator)
	return "/" + path, nil
}

// LookupSubdirectory will return a directory for the specified subdirectory relative to this
// directory.  It assumes that subdirectory is a relative path, even if it begins with a path
// separator character.  If the specified subdirectory can't be found, or if any named directory
// entry along its path is not a directory (e.g. if it is a file), then it will return an error
func (d *directory) LookupSubdirectory(subdirectory string) (Directory, error) {
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(subdirectory)
	if err != nil {
		return nil, err
	}
	return NewDirectory(subdirInode), nil
}

func (d *directory) Mkdir(subdirectory string) (Directory, error) {
	pathInfo := filepath.ParsePath(subdirectory)
	if !pathInfo.IsRelative {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the directory that will be parent to the subdirectory
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(pathInfo.ParentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create %s", subdirectory)
	}
	// Create the directory
	newDirInode, err := subdirInode.AddDirectory(pathInfo.Entry)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create %s", subdirectory)
	}
	return NewDirectory(newDirInode), nil
}

func (d *directory) ReadDir(subdirectory string) ([]DirectoryEntry, error) {
	// Validate that the path is relative
	if !filepath.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the DirectoryInode for the subdirectory
	dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list entries in '%s'", subdirectory)
	}
	// Get the directory inode entries
	inodeEntries := dirInode.InodeEntries()
	toReturn := make([]DirectoryEntry, 0, len(inodeEntries))
	for _, entry := range inodeEntries {
		toReturn = append(toReturn, DirectoryEntry{
			Name: entry.Name,
			Type: directoryEntryTypeFromInodeType(entry.Type),
		})
	}
	return toReturn, nil
}

func (d *directory) Rmdir(subdirectory string) error {
	pathInfo := filepath.ParsePath(subdirectory)
	if !pathInfo.IsRelative {
		return fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the directory that is parent to the subdirectory
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(pathInfo.ParentPath)
	if err != nil {
		return errors.Wrapf(err, "could not delete '%s'", subdirectory)
	}
	// Remove the directory
	if err := subdirInode.DeleteDirectory(pathInfo.Entry); err != nil {
		return errors.Wrapf(err, "could not delete '%s'", subdirectory)
	}
	return nil
}

func (d *directory) CreateFile(relativePath string) (file.File, error) {
	f, err := d.OpenFile(relativePath, modes.OpenFileModeEqualToCreateFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create '%s'", relativePath)
	}
	return f, nil
}

func (d *directory) OpenFile(relativePath string, mode int) (file.File, error) {
	pathInfo := filepath.ParsePath(relativePath)
	if !pathInfo.IsRelative {
		return nil, fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	if pathInfo.MustBeDir {
		return nil, errors.Wrapf(fserrors.EInval, "path specifies a directory")
	}
	// Lookup the directory that is parent to the relativePath
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(pathInfo.ParentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open '%s'", relativePath)
	}
	// Get the file
	var fileInode *inode.FileInode
	if modes.IsCreateMode(mode) {
		fileInode, err = subdirInode.CreateFileInodeEntry(pathInfo.Entry, modes.IsExclusiveMode(mode))
	} else {
		fileInode, err = subdirInode.FileInodeEntry(pathInfo.Entry)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "could not open %s", relativePath)
	}
	// Truncate the file if the mode says to do so
	if modes.IsTruncateMode(mode) {
		err := fileInode.TruncateAndWriteAll(make([]byte, 0))
		if err != nil {
			return nil, errors.Wrapf(err, "could not truncate %s on open", relativePath)
		}
	}
	return file.NewFile(fileInode, mode), nil
}

func (d *directory) Stat(relativePath string) (*FileInfo, error) {
	pathInfo := filepath.ParsePath(relativePath)
	if !pathInfo.IsRelative {
		return nil, fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	// Lookup the directory that is parent to the relativePath
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(pathInfo.ParentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not stat '%s'", relativePath)
	}
	// Grab the file or directory inode from subdirInode
	genericInode, err := subdirInode.InodeEntry(pathInfo.Entry)
	if err != nil {
		return nil, errors.Wrapf(err, "could not stat %s", relativePath)
	}
	switch inodeTyped := genericInode.(type) {
	case *inode.FileInode:
		if pathInfo.MustBeDir {
			return nil, errors.Wrapf(fserrors.ENotDir, "file found where directory %s expected", relativePath)
		}
		return &FileInfo{
			Type: FileType,
			Size: inodeTyped.Size(),
		}, nil
	case *inode.DirectoryInode:
		return &FileInfo{
			Type: DirectoryType,
			Size: inodeTyped.Size(),
		}, nil
	default:
		return nil, fmt.Errorf("malformed inoded of type '%s' on path '%s'", genericInode.InodeType().String(), relativePath)
	}
}

func (d *directory) DeleteFile(relativePath string) error {
	pathInfo := filepath.ParsePath(relativePath)
	if !pathInfo.IsRelative {
		return fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	if pathInfo.MustBeDir {
		return errors.Wrapf(fserrors.EInval, "path specifies a directory")
	}
	// Lookup the directory that will be parent to the relativePath
	subdirInode, err := d.DirectoryInode.LookupSubdirectory(pathInfo.ParentPath)
	if err != nil {
		return errors.Wrapf(err, "could not delete '%s'", relativePath)
	}
	// Remove the file
	if err := subdirInode.DeleteFile(pathInfo.Entry); err != nil {
		return errors.Wrapf(err, "could not delete '%s'", relativePath)
	}
	return nil
}

// Parse parent
func (d *directory) Rename(srcRelativePath, dstRelativePath string) error {
	srcPathInfo := filepath.ParsePath(srcRelativePath)
	dstPathInfo := filepath.ParsePath(dstRelativePath)
	// Validate that both parts are relative
	if !srcPathInfo.IsRelative {
		return fmt.Errorf("'%s' is not a relative path", srcRelativePath)
	}
	if !dstPathInfo.IsRelative {
		return fmt.Errorf("'%s' is not a relative path", dstRelativePath)
	}
	// Look up the directories that are parent to src and dst
	srcDirInode, err := d.DirectoryInode.LookupSubdirectory(srcPathInfo.ParentPath)
	if err != nil {
		return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
	}
	dstDirInode, err := d.DirectoryInode.LookupSubdirectory(dstPathInfo.ParentPath)
	if err != nil {
		return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
	}
	// Move the entry
	if err := inode.MoveEntry(srcDirInode, dstDirInode, srcPathInfo, dstPathInfo); err != nil {
		return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
	}
	return nil
}
