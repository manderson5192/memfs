package directory

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/inode"
	"github.com/manderson5192/memfs/utils"
	"github.com/pkg/errors"
)

const (
	SelfDirectoryEntry   = inode.SelfDirectoryEntry
	ParentDirectoryEntry = inode.ParentDirectoryEntry
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
	// OpenFile returns a reference to the file at the specified relative path, or returns an error
	OpenFile(relativePath string) (file.File, error)
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
	// Validate that the path is relative and non-empty
	if subdirectory == "" {
		return nil, fmt.Errorf("no subdirectory provided")
	}
	if !filepath.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the directory that will be parent to the subdirectory
	subdirNameToLookup, dirNameToCreate, found := utils.RightCut(subdirectory, filepath.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create %s", subdirectory)
		}
		subdirInode = dirInode
	}
	// Create the directory
	newDirInode, err := subdirInode.AddDirectory(dirNameToCreate)
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
	// Validate that the path is relative and non-empty
	if subdirectory == "" {
		return fmt.Errorf("no subdirectory provided")
	}
	if !filepath.IsRelativePath(subdirectory) {
		return fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the directory from which the named subdirectory will be removed
	subdirNameToLookup, dirNameToDelete, found := utils.RightCut(subdirectory, filepath.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
		if err != nil {
			return errors.Wrapf(err, "could not delete '%s'", subdirectory)
		}
		subdirInode = dirInode
	}
	// Remove the directory
	if err := subdirInode.DeleteDirectory(dirNameToDelete); err != nil {
		return errors.Wrapf(err, "could not delete '%s'", subdirectory)
	}
	return nil
}

func (d *directory) CreateFile(relativePath string) (file.File, error) {
	// Validate that the path is relative and non-empty
	if relativePath == "" {
		return nil, fmt.Errorf("no path provided")
	}
	if !filepath.IsRelativePath(relativePath) {
		return nil, fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	// Lookup the directory that will be parent to the file
	subdirNameToLookup, fileToCreate, found := utils.RightCut(relativePath, filepath.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create %s", relativePath)
		}
		subdirInode = dirInode
	}
	// Create the file
	newFileInode, err := subdirInode.AddFile(fileToCreate)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create %s", relativePath)
	}
	return file.NewFile(newFileInode), nil
}

func (d *directory) OpenFile(relativePath string) (file.File, error) {
	// Validate that the path is relative and non-empty
	if relativePath == "" {
		return nil, fmt.Errorf("no path provided")
	}
	if !filepath.IsRelativePath(relativePath) {
		return nil, fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	// Lookup the directory that is parent to the file
	subdirNameToLookup, filename, found := utils.RightCut(relativePath, filepath.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
		if err != nil {
			return nil, errors.Wrapf(err, "could not open %s", relativePath)
		}
		subdirInode = dirInode
	}
	// Get the file
	fileInode, err := subdirInode.FileInodeEntry(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open %s", relativePath)
	}
	return file.NewFile(fileInode), nil
}

func (d *directory) Stat(relativePath string) (*FileInfo, error) {
	genericInode := inode.Inode(d.DirectoryInode)
	if relativePath != "" {
		// Validate that the path is relative
		if !filepath.IsRelativePath(relativePath) {
			return nil, fmt.Errorf("'%s' is not a relative path", relativePath)
		}
		// Lookup the directory that is parent to the file
		subdirNameToLookup, filename, found := utils.RightCut(relativePath, filepath.PathSeparator)
		subdirInode := d.DirectoryInode
		if found {
			dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
			if err != nil {
				return nil, errors.Wrapf(err, "could not stat %s", relativePath)
			}
			subdirInode = dirInode
		}
		var err error
		genericInode, err = subdirInode.InodeEntry(filename)
		if err != nil {
			return nil, errors.Wrapf(err, "could not stat %s", relativePath)
		}
	}
	switch inodeTyped := genericInode.(type) {
	case *inode.FileInode:
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
	// Validate that the path is relative and non-empty
	if relativePath == "" {
		return fmt.Errorf("no path provided")
	}
	if !filepath.IsRelativePath(relativePath) {
		return fmt.Errorf("'%s' is not a relative path", relativePath)
	}
	// Lookup the directory from which the named subdirectory will be removed
	subdirNameToLookup, fileToDelete, found := utils.RightCut(relativePath, filepath.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(subdirNameToLookup)
		if err != nil {
			return errors.Wrapf(err, "could not delete '%s'", relativePath)
		}
		subdirInode = dirInode
	}
	// Remove the file
	if err := subdirInode.DeleteFile(fileToDelete); err != nil {
		return errors.Wrapf(err, "could not delete '%s'", relativePath)
	}
	return nil
}

func (d *directory) Rename(srcRelativePath, dstRelativePath string) error {
	// Validate that both paths are relative and non-empty
	if srcRelativePath == "" {
		return fmt.Errorf("no source path provided")
	}
	if !filepath.IsRelativePath(srcRelativePath) {
		return fmt.Errorf("'%s' is not a relative path", srcRelativePath)
	}
	if dstRelativePath == "" {
		return fmt.Errorf("no destination path provided")
	}
	if !filepath.IsRelativePath(dstRelativePath) {
		return fmt.Errorf("'%s' is not a relative path", dstRelativePath)
	}
	// Parse the parent directory and entry name from srcRelativePath
	srcDirToLookup, srcEntryToMove, found := utils.RightCut(srcRelativePath, filepath.PathSeparator)
	srcDirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(srcDirToLookup)
		if err != nil {
			return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
		}
		srcDirInode = dirInode
	}
	// ...now do the same for dstRelativePath
	dstDirToLookup, dstEntryToMove, found := utils.RightCut(dstRelativePath, filepath.PathSeparator)
	dstDirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.LookupSubdirectory(dstDirToLookup)
		if err != nil {
			return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
		}
		dstDirInode = dirInode
	}
	// Move the entry
	if err := inode.MoveEntry(srcDirInode, dstDirInode, srcEntryToMove, dstEntryToMove); err != nil {
		return errors.Wrapf(err, "could not rename '%s' to '%s'", srcRelativePath, dstRelativePath)
	}
	return nil
}
