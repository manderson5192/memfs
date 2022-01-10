package inode

import (
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/filepath"
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
	defer i.rwMutex.RUnlock()
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
	if strings.Contains(name, filepath.PathSeparator) {
		return nil, fmt.Errorf("cannot add subdirectory inode for a name containing path separator %s: %s", filepath.PathSeparator, name)
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Disallow adding subdirectories on directories that have already been marked as deleted
	if i.deleted {
		return nil, fmt.Errorf("cannot add entries to a directory marked for deletion")
	}
	// Make sure that the entry doesn't already exist
	if _, exists := i.contents[name]; exists {
		return nil, fmt.Errorf("directory entry '%s' already exists", name)
	}
	subdirInode := NewDirectoryInode(i)
	i.contents[name] = subdirInode
	return subdirInode, nil
}

// AddFile adds (and returns) a FileInode for a direct child file named 'name'.  It cannot create a
// with a name containing the path separator and it cannot create a file whose name is already taken
func (i *DirectoryInode) AddFile(name string) (*FileInode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(name, filepath.PathSeparator) {
		return nil, fmt.Errorf("cannot add file inode for a name containing path separator %s: %s", filepath.PathSeparator, name)
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Disallow adding files to directories that have already been marked as deleted
	if i.deleted {
		return nil, fmt.Errorf("cannot add entries to a directory marked for deletion")
	}
	// Make sure that the entry doesn't already exist
	if _, exists := i.contents[name]; exists {
		return nil, fmt.Errorf("directory entry '%s' already exists", name)
	}
	fileInode := NewFileInode()
	i.contents[name] = fileInode
	return fileInode, nil
}

// DirectoryInodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) DirectoryInodeEntry(entry string) (*DirectoryInode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(entry, filepath.PathSeparator) {
		return nil, fmt.Errorf("entry %s contains illegal character %s", entry, filepath.PathSeparator)
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

// FileInodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) FileInodeEntry(entry string) (*FileInode, error) {
	// Check that this entry doesn't contain the path separator
	if strings.Contains(entry, filepath.PathSeparator) {
		return nil, fmt.Errorf("entry %s contains illegal character %s", entry, filepath.PathSeparator)
	}
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	inode, exists := i.contents[entry]
	if !exists {
		return nil, fmt.Errorf("entry '%s' does not exist", entry)
	}
	fileInode, ok := inode.(*FileInode)
	if !ok {
		return nil, fmt.Errorf("entry '%s' is not a file", entry)
	}
	return fileInode, nil
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
	if !filepath.IsRelativePath(subdirectory) {
		return nil, fmt.Errorf("'%s' is not a relative path", subdirectory)
	}
	currentDirInode := i
	currentSubdirectory := subdirectory
	for len(currentSubdirectory) > 0 {
		// Parse a directory entry from the beginning of currentSubdirectory
		currentSubdirectory = strings.TrimLeft(currentSubdirectory, filepath.PathSeparator)
		entryName, remainder, _ := utils.Cut(currentSubdirectory, filepath.PathSeparator)
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
	for entry := range i.contents {
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

// this function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doDeleteDirectory(entry string) error {
	// Check: disallow removing the special "." and ".." directories
	if entry == "." || entry == ".." {
		return fmt.Errorf("refusing to remove '.' or '..' directory: skipping '%s", entry)
	}
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

func (i *DirectoryInode) DeleteDirectory(entry string) error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	return i.doDeleteDirectory(entry)
}

// this function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doDeleteFile(entry string) error {
	// Get the FileInode for entry
	inode, exists := i.contents[entry]
	if !exists {
		return fmt.Errorf("entry '%s' does not exist", entry)
	}
	_, ok := inode.(*FileInode)
	if !ok {
		return fmt.Errorf("entry '%s' is not a file", entry)
	}
	// Remove the entry
	delete(i.contents, entry)
	return nil
}

func (i *DirectoryInode) DeleteFile(entry string) error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	return i.doDeleteFile(entry)
}

// this function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doInsertFileInode(entry string, newEntry *FileInode) error {
	// if an entry by this name already exists, then we are meant to delete it
	if oldEntry, exists := i.contents[entry]; exists {
		switch oldEntry.(type) {
		case *FileInode:
			if err := i.doDeleteFile(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing file")
			}
		case *DirectoryInode:
			if err := i.doDeleteDirectory(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing directory")
			}
		default:
			return fmt.Errorf("existing entry '%s' has malformed inode of type '%s'", entry, oldEntry.InodeType().String())
		}
	}
	i.contents[entry] = newEntry
	return nil
}

// this function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doInsertDirectoryInode(entry string, newEntry *DirectoryInode) error {
	// if an entry by this name already exists, then we are meant to delete it
	if oldEntry, exists := i.contents[entry]; exists {
		switch oldEntry.(type) {
		case *FileInode:
			// Interestingly, the POSIX spec says that rename(2) should return an error (EISDIR)
			// if the source ("old") path specifies a directory but the destination ("new") path
			// coincides with a file.  We could do that here, but it doesn't seem strictly
			// necessary, so we will allow it.
			if err := i.doDeleteFile(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing file")
			}
		case *DirectoryInode:
			if err := i.doDeleteDirectory(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing directory")
			}
		default:
			return fmt.Errorf("existing entry '%s' has malformed inode of type '%s'", entry, oldEntry.InodeType().String())
		}
	}
	// insert the entry into this directory
	i.contents[entry] = newEntry
	// update the newEntry inode's parent pointer to point to this inode
	newEntry.SetParent(i)
	return nil
}

func (i *DirectoryInode) SetParent(parent *DirectoryInode) {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	i.contents[ParentDirectoryEntry] = parent
}

func MoveEntry(srcParentInode, dstParentInode *DirectoryInode, srcEntry, dstEntry string) error {
	// Check that srcEntry is not the special self or parent directory entries
	if srcEntry == SelfDirectoryEntry || srcEntry == ParentDirectoryEntry {
		return fmt.Errorf("cannot move '.' or '..' entries")
	}
	// Check the same for dstEntry
	if dstEntry == SelfDirectoryEntry || dstEntry == ParentDirectoryEntry {
		return fmt.Errorf("cannot overwrite '.' or '..' entries")
	}
	// Check that the dst entry name doesn't contain the path separator
	if strings.Contains(dstEntry, filepath.PathSeparator) {
		return fmt.Errorf("entry name '%s' contains the path separator", dstEntry)
	}
	// Edge case: srcParentInode and dstParentInode are the same.  That requires a different locking
	// discipline, so we special-case it
	if srcParentInode == dstParentInode {
		return srcParentInode.renameEntry(srcEntry, dstEntry)
	}
	srcParentInode.rwMutex.Lock()
	defer srcParentInode.rwMutex.Unlock()
	dstParentInode.rwMutex.Lock()
	defer dstParentInode.rwMutex.Unlock()
	// Disallow adding files to directories that have already been marked as deleted
	if dstParentInode.deleted {
		return fmt.Errorf("cannot add entries to a directory marked for deletion")
	}
	// Get the inode for the srcEntry
	srcInode, exists := srcParentInode.contents[srcEntry]
	if !exists {
		return fmt.Errorf("source entry '%s' does not exist", srcEntry)
	}
	// Insert the inode into its new location
	switch srcInodeTyped := srcInode.(type) {
	case *FileInode:
		if err := dstParentInode.doInsertFileInode(dstEntry, srcInodeTyped); err != nil {
			return err
		}
	case *DirectoryInode:
		if err := dstParentInode.doInsertDirectoryInode(dstEntry, srcInodeTyped); err != nil {
			return err
		}
	default:
		return fmt.Errorf("source entry '%s' has malformed inode of type '%s'", srcEntry, srcInode.InodeType().String())
	}
	// Remove the inode from its old location
	delete(srcParentInode.contents, srcEntry)
	return nil
}

func (i *DirectoryInode) renameEntry(srcEntry, dstEntry string) error {
	// Special case: do nothing
	if srcEntry == dstEntry {
		return nil
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Disallow moving files in directories that have already been marked as deleted
	if i.deleted {
		return fmt.Errorf("cannot move entries in a directory marked for deletion")
	}
	// Get the inode to be moved
	inode, exists := i.contents[srcEntry]
	if !exists {
		return fmt.Errorf("source entry '%s' does not exist", srcEntry)
	}
	switch inodeTyped := inode.(type) {
	case *FileInode:
		if err := i.doInsertFileInode(dstEntry, inodeTyped); err != nil {
			return err
		}
	case *DirectoryInode:
		if err := i.doInsertDirectoryInode(dstEntry, inodeTyped); err != nil {
			return err
		}
	default:
		return fmt.Errorf("source entry '%s' has malformed inode of type '%s'", srcEntry, inodeTyped.InodeType().String())
	}
	delete(i.contents, srcEntry)
	return nil
}
