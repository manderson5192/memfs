package directory

import (
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/inode"
	"github.com/manderson5192/memfs/path"
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

// DirectoryEntry represents a file or directory entry in a given directory
type DirectoryEntry struct {
	// Name is the entry's name
	Name string
	// Type indicates whether the entry is a file or a directory
	Type DirectoryEntryType
}

type Directory interface {
	// Equals returns true if the other Directory references the same inode, false otherwise
	Equals(other Directory) bool
	// ReversePathLookup returns a valid absolute path for the directory or an error
	ReversePathLookup() (string, error)
	// Lookup returns the Directory for the subdirectory of the current directory or an error
	Lookup(subdirectory string) (Directory, error)
	// Mkdir creates and returns a Directory for the specified subdirectory of the current
	// directory, or returns an error.  It will return an error if a path component does not exist
	// or is not a directory.  It will return an error if the specified subdirectory already exists.
	Mkdir(subdirectory string) (Directory, error)
	// ReadDir returns an array of DirectoryEntry for the specified subdirectory of the current
	// directory, or returns an error.  It will return an error if a path component does not exist
	// or is not a directory.
	ReadDir(subdirectory string) ([]DirectoryEntry, error)
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
	// TODO: what if currentDirInode is deleted?  Is that even possible?
	for !currentDirInode.IsRootDirectoryInode() {
		parentDirInode := currentDirInode.Parent()
		pathPart, err := parentDirInode.ReverseLookupEntry(currentDirInode)
		if err != nil {
			return "", errors.Wrapf(err, "could not complete reverse path lookup")
		}
		pathParts = append([]string{pathPart}, pathParts...)
		currentDirInode = parentDirInode
	}
	path := strings.Join(pathParts, path.PathSeparator)
	return "/" + path, nil
}

// Lookup will return a directory for the specified subdirectory relative to this directory.  It
// assumes that subdirectory is a relative path, even if it begins with a path separator character.
// If the specified subdirectory can't be found, or if any named directory entry along its path is
// not a directory (e.g. if it is a file), then it will return an error
func (d *directory) Lookup(subdirectory string) (Directory, error) {
	subdirInode, err := d.DirectoryInode.Lookup(subdirectory)
	if err != nil {
		return nil, err
	}
	return NewDirectory(subdirInode), nil
}

func (d *directory) Mkdir(subdirectory string) (Directory, error) {
	// Validate that the path is relative
	if !path.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the directory that will be parent to the subdirectory
	subdirNameToLookup, dirNameToCreate, found := utils.RightCut(subdirectory, path.PathSeparator)
	subdirInode := d.DirectoryInode
	if found {
		dirInode, err := d.DirectoryInode.Lookup(subdirNameToLookup)
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
	if !path.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	// Lookup the DirectoryInode for the subdirectory
	dirInode, err := d.DirectoryInode.Lookup(subdirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list entries in '%s'", subdirectory)
	}
	// Get the directory inode entries
	inodeEntries := dirInode.DirectoryEntries()
	toReturn := make([]DirectoryEntry, 0, len(inodeEntries))
	for _, entry := range inodeEntries {
		toReturn = append(toReturn, DirectoryEntry{
			Name: entry.Name,
			Type: directoryEntryTypeFromInodeType(entry.Type),
		})
	}
	return toReturn, nil
}
