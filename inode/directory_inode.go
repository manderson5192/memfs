package inode

import (
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/path"
	"github.com/manderson5192/memfs/utils"
	"github.com/pkg/errors"
)

const (
	SelfDirectoryEntry   string = "."
	ParentDirectoryEntry string = ".."
)

type DirectoryInode struct {
	basicInode
	contents map[string]Inode
}

func NewRootDirectoryInode() *DirectoryInode {
	rootDirInode := &DirectoryInode{
		contents: map[string]Inode{},
	}
	rootDirInode.contents[SelfDirectoryEntry] = rootDirInode
	rootDirInode.contents[ParentDirectoryEntry] = rootDirInode
	return rootDirInode
}

func NewDirectoryInode(parent *DirectoryInode) *DirectoryInode {
	newDirInode := &DirectoryInode{
		contents: map[string]Inode{},
	}
	newDirInode.contents[SelfDirectoryEntry] = newDirInode
	newDirInode.contents[ParentDirectoryEntry] = parent
	return newDirInode
}

func (i *DirectoryInode) InodeType() InodeType {
	return InodeDirectory
}

// Parent obtains the DirectoryInode that is parent to this DirectoryInode
func (i *DirectoryInode) Parent() *DirectoryInode {
	i.rwMutex.RLock()
	defer i.rwMutex.RLocker().Unlock()
	parentInode, parentExists := i.contents[ParentDirectoryEntry]
	if !parentExists {
		// This shouldn't happen, so we panic on the condition
		panic("parent entry for directory inode does not exist")
	}
	parentDirectoryInode, castOk := parentInode.(*DirectoryInode)
	if !castOk {
		// This also shouldn't happen
		panic("parent directory cannot cast to directory inode type")
	}
	return parentDirectoryInode
}

// ReverseLookupEntry returns the entry name for the specified child DirectoryInode, or an error
// if it is unable to do so
func (i *DirectoryInode) ReverseLookupEntry(child *DirectoryInode) (string, error) {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	for entry, inode := range i.contents {
		// Ignore self and parent directory references
		if entry == SelfDirectoryEntry || entry == ParentDirectoryEntry {
			continue
		}
		// Ignore non-directory inodes
		if !IsDirectory(inode) {
			continue
		}
		if child == inode {
			return entry, nil
		}
	}
	return "", fmt.Errorf("entry for directory inode was not found")
}

// IsRootDirectoryInode returns whether this DirectoryInode corresponds to the filesystem's root
func (i *DirectoryInode) IsRootDirectoryInode() bool {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	parent := i.Parent()
	return i == parent
}

// AddDirectory adds (and returns) a DirectoryInode for a direct child directory named 'name'.  It
// cannot create an entry containing a path separator and it cannot create a subdirectory that
// already exists
func (i *DirectoryInode) AddDirectory(name string) (*DirectoryInode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(name, path.PathSeparator) {
		return nil, fmt.Errorf("cannot add subdirectory inode for a name containing path separator %s: %s", path.PathSeparator, name)
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Make sure that the entry doesn't already exist
	if _, exists := i.contents[name]; exists {
		return nil, fmt.Errorf("subdirectory entry '%s' already exists", name)
	}
	subdirInode := NewDirectoryInode(i)
	i.contents[name] = subdirInode
	return subdirInode, nil
}

// DirectoryEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) DirectoryEntry(entry string) (Inode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(entry, path.PathSeparator) {
		return nil, fmt.Errorf("entry %s contains illegal character %s", entry, path.PathSeparator)
	}
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	inode, exists := i.contents[entry]
	if !exists {
		return nil, fmt.Errorf("entry does not exist: '%s'", entry)
	}
	return inode, nil
}

type DirectoryInodeEntry struct {
	Name string
	Type InodeType
}

func (i *DirectoryInode) DirectoryEntries() []DirectoryInodeEntry {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	toReturn := make([]DirectoryInodeEntry, 0, len(i.contents))
	for entryName, inode := range i.contents {
		if entryName == SelfDirectoryEntry || entryName == ParentDirectoryEntry {
			continue
		}
		toReturn = append(toReturn, DirectoryInodeEntry{
			Name: entryName,
			Type: inode.InodeType(),
		})
	}
	return toReturn
}

// Lookup will return a DirectoryInode for the specified subdirectory relative to this
// DirectoryInode.  It assumes that subdirectory is a relative path, even if it begins with a path
// separator character.  If the specified subdirectory can't be found, or if any named directory
// entry along its path is not a directory (e.g. if it is a file), then it will return an error
func (i *DirectoryInode) Lookup(subdirectory string) (*DirectoryInode, error) {
	if !path.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	currentDirInode := i
	currentSubdirectory := subdirectory
	for len(currentSubdirectory) > 0 {
		// Parse a directory entry from the beginning of currentSubdirectory
		currentSubdirectory = strings.TrimLeft(currentSubdirectory, path.PathSeparator)
		entryName, remainder, _ := utils.Cut(currentSubdirectory, path.PathSeparator)
		// Get the directory inode for this entry
		inode, getEntryErr := currentDirInode.DirectoryEntry(entryName)
		if getEntryErr != nil {
			return nil, errors.Wrapf(getEntryErr, "subdirectory '%s' does not exist", subdirectory)
		}
		dirInode, ok := inode.(*DirectoryInode)
		if !ok {
			return nil, fmt.Errorf("subdirectory %s not found: %s is not a directory", subdirectory, entryName)
		}
		// iterate
		currentDirInode = dirInode
		currentSubdirectory = remainder
	}
	return currentDirInode, nil
}
