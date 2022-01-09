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
	deleted  bool
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
	// Disallow adding subdirectories on directories that have already been marked as deleted
	if i.deleted {
		return nil, fmt.Errorf("cannot add subdirectories to a directory marked for deletion")
	}
	// Make sure that the entry doesn't already exist
	if _, exists := i.contents[name]; exists {
		return nil, fmt.Errorf("subdirectory entry '%s' already exists", name)
	}
	subdirInode := NewDirectoryInode(i)
	i.contents[name] = subdirInode
	return subdirInode, nil
}

// InodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) InodeEntry(entry string) (Inode, error) {
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

// InodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) DirectoryInodeEntry(entry string) (*DirectoryInode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(entry, path.PathSeparator) {
		return nil, fmt.Errorf("entry %s contains illegal character %s", entry, path.PathSeparator)
	}
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	inode, exists := i.contents[entry]
	if !exists {
		return nil, fmt.Errorf("entry '%s' does not exist", entry)
	}
	dirInode, ok := inode.(*DirectoryInode)
	if !ok {
		return nil, fmt.Errorf("entry '%s' is not a directory", entry)
	}
	// Deny access to DirectoryInodes after they have been marked as deleted.  This case should be
	// rare, but is technically possible
	if dirInode.isDeleted() {
		return nil, fmt.Errorf("entry '%s' does not exist", entry)
	}
	return dirInode, nil
}

type InodeEntry struct {
	Name string
	Type InodeType
}

func (i *DirectoryInode) InodeEntries() []InodeEntry {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	toReturn := make([]InodeEntry, 0, len(i.contents))
	for entryName, inode := range i.contents {
		if entryName == SelfDirectoryEntry || entryName == ParentDirectoryEntry {
			continue
		}
		toReturn = append(toReturn, InodeEntry{
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
func (i *DirectoryInode) LookupSubdirectory(subdirectory string) (*DirectoryInode, error) {
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
		dirInode, getEntryErr := currentDirInode.DirectoryInodeEntry(entryName)
		if getEntryErr != nil {
			return nil, errors.Wrapf(getEntryErr, "cannot find subdirectory '%s'", subdirectory)
		}
		// iterate
		currentDirInode = dirInode
		currentSubdirectory = remainder
	}
	return currentDirInode, nil
}

// delete marks this DirectoryInode as deleted.  It will only succeed if this directory is empty.
func (i *DirectoryInode) delete() error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Check: is the directory already deleted?
	if i.deleted {
		return nil
	}
	// Check: is the directory empty?
	for entry, _ := range i.contents {
		if entry == SelfDirectoryEntry || entry == ParentDirectoryEntry {
			continue
		}
		return fmt.Errorf("directory is not empty")
	}
	// mark as deleted
	i.deleted = true
	return nil
}

func (i *DirectoryInode) isDeleted() bool {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	return i.deleted
}

func (i *DirectoryInode) DeleteDirectory(entry string) error {
	// Check: disallow removing the special "." and ".." directories
	if entry == "." || entry == ".." {
		return fmt.Errorf("refusing to remove '.' or '..' directory: skipping '%s", entry)
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Get the DirectoryInode for entry
	inode, exists := i.contents[entry]
	if !exists {
		return fmt.Errorf("entry '%s' does not exist", entry)
	}
	dirInode, ok := inode.(*DirectoryInode)
	if !ok {
		return fmt.Errorf("entry '%s' is not a directory", entry)
	}
	// Make sure we can successfully delete entry's directory
	if err := dirInode.delete(); err != nil {
		return errors.Wrapf(err, "failed to delete directory entry '%s'", entry)
	}
	// Finally, remove the entry
	delete(i.contents, entry)
	return nil
}
